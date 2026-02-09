package client

import (
	"autoSSL/ent/user"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"

	"autoSSL/ent"
	"autoSSL/ent/acmeaccount"

	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// getCADirectoryURL 获取不同CA的目录URL
func getCADirectoryURL(provider string) string {
	switch provider {
	case "letsencrypt":
		return lego.LEDirectoryProduction
	case "letsencrypt-staging":
		return lego.LEDirectoryStaging
	case "zerossl":
		return "https://acme.zerossl.com/v2/DV90"
	default:
		return lego.LEDirectoryProduction
	}
}

// ACMEAccountManager 管理ACME账号的创建和复用
type ACMEAccountManager struct {
	client      *ent.Client
	rateLimiter *RateLimiter
}

// NewACMEAccountManager 创建新的ACME账号管理器
func NewACMEAccountManager(client *ent.Client, rateLimiter *RateLimiter) *ACMEAccountManager {
	return &ACMEAccountManager{
		client:      client,
		rateLimiter: rateLimiter,
	}
}

// GetOrCreateACMEAccount 根据用户ID和邮箱获取或创建ACME账号（保持向后兼容）
// 默认使用Let's Encrypt作为证书提供商
// func (m *ACMEAccountManager) GetOrCreateACMEAccount(ctx context.Context, userID int, email string) (Client, error) {
// 	return m.GetOrCreateAccountWithProvider(ctx, userID, email, string(ProviderLetsEncrypt))
// }

// GetOrCreateAccountWithProvider 根据用户ID、邮箱和CA提供者获取或创建账号
// userID: 用户ID
// email: 用户邮箱地址
// provider: 证书提供商（letsencrypt, zerossl）
// 返回配置好的SSL客户端
func (m *ACMEAccountManager) GetOrCreateAccountWithProvider(ctx context.Context, userID int, email string, provider string) (Client, error) {
	// 查询该用户已存在的账号
	account, err := m.client.ACMEAccount.Query().
		Where(
			acmeaccount.Email(email),
			acmeaccount.Provider(provider),
			acmeaccount.HasUserWith(user.ID(userID)),
		).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, fmt.Errorf("查询ACME账号失败: %v", err)
	}

	// 日志记录（可选，生产环境可以移除）
	// fmt.Printf("用户id是：%d\n", userID)

	// 如果存在，使用现有账号
	if account != nil {
		return m.createClientFromAccountWithProvider(account, provider)
	}

	// 检查用户是否允许创建新ACME账户
	userIDStr := strconv.Itoa(userID)
	if allowed, waitTime := m.rateLimiter.AllowAccountCreation(userIDStr); !allowed {
		return nil, fmt.Errorf("用户ACME账户创建受速率限制，需要等待: %v", waitTime)
	}

	// 创建新的ACME账户并记录
	client, err := m.createNewACMEAccountWithProvider(ctx, userID, email, provider)
	if err == nil {
		m.rateLimiter.RecordAccountCreation(userIDStr)
	}
	return client, err
}

// createClientFromAccountWithProvider 从数据库中的账号信息创建指定CA的ACME客户端
func (m *ACMEAccountManager) createClientFromAccountWithProvider(account *ent.ACMEAccount, provider string) (Client, error) {
	// 解析私钥
	block, _ := pem.Decode([]byte(account.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("无效的私钥格式")
	}

	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %v", err)
	}

	// 解析注册信息
	var reg registration.Resource
	if err := json.Unmarshal([]byte(account.RegistrationBody), &reg); err != nil {
		return nil, fmt.Errorf("解析注册信息失败: %v", err)
	}

	// 创建ACME用户
	myUser := &acmeUser{
		Email:        account.Email,
		key:          privateKey,
		registration: &reg,
	}

	// 创建lego客户端
	config := lego.NewConfig(myUser)
	config.CADirURL = getCADirectoryURL(provider)

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建lego客户端失败: %v", err)
	}

	return &ACMEClient{
		client: client,
		email:  account.Email,
	}, nil
}

// createNewACMEAccountWithProvider 创建指定CA的新ACME账号并保存到数据库
func (m *ACMEAccountManager) createNewACMEAccountWithProvider(ctx context.Context, userID int, email string, provider string) (Client, error) {
	// 验证用户是否存在
	_, err := m.client.User.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在，无法创建ACME账号: %v", err)
	}

	// 生成新的私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成私钥失败: %v", err)
	}

	// 创建ACME用户
	myUser := &acmeUser{
		Email: email,
		key:   privateKey,
	}

	// 创建lego客户端
	config := lego.NewConfig(myUser)
	config.CADirURL = getCADirectoryURL(provider)

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建lego客户端失败: %v", err)
	}

	// 注册账号
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("注册ACME账号失败: %v", err)
	}
	myUser.registration = reg

	// 序列化私钥
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("序列化私钥失败: %v", err)
	}
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 序列化注册信息
	regBody, err := json.Marshal(reg)
	if err != nil {
		return nil, fmt.Errorf("序列化注册信息失败: %v", err)
	}

	// 保存到数据库
	_, err = m.client.ACMEAccount.Create().
		SetEmail(email).
		SetProvider(provider).
		SetPrivateKey(string(privateKeyPEM)).
		SetRegistrationURI(reg.URI).
		SetRegistrationBody(string(regBody)).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("保存ACME账号到数据库失败: %v", err)
	}

	return &ACMEClient{
		client: client,
		email:  email,
	}, nil
}
