package client

import (
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

// ACMEClient ACME协议客户端实现
// 实现了Client接口，支持Let's Encrypt和ZeroSSL等ACME兼容的证书颁发机构
type ACMEClient struct {
	client *lego.Client
	email  string
}

// NewACMEClientWithUser 使用已存在的ACME用户信息创建客户端
// user: ACME用户信息，包含邮箱、私钥和注册信息
// caDirectoryURL: 证书颁发机构的ACME目录URL
// 返回配置好的ACME客户端实例
func NewACMEClientWithUser(user *acmeUser, caDirectoryURL string) (*ACMEClient, error) {
	config := lego.NewConfig(user)
	config.CADirURL = caDirectoryURL

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建lego客户端失败: %v", err)
	}

	return &ACMEClient{
		client: client,
		email:  user.Email,
	}, nil
}

type acmeUser struct {
	Email        string
	key          crypto.PrivateKey
	registration *registration.Resource
}

func (u *acmeUser) GetEmail() string {
	return u.Email
}

func (u *acmeUser) GetRegistration() *registration.Resource {
	return u.registration
}

func (u *acmeUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// RequestCertificate 请求SSL证书
// domain: 证书域名
// dnsProvider: DNS提供者信息结构体，包含ProviderType、APIKey和APISecret
// 返回证书内容、私钥和可能的错误信息
func (c *ACMEClient) RequestCertificate(domain string, dnsProvider CertificateRequestInfo) (string, string, error) {
	// 设置DNS提供者
	factory := NewDNSProviderFactory(dnsProvider.APIKey, dnsProvider.APISecret, dnsProvider.APIToken)
	provider, err := factory.CreateProvider(dnsProvider.ProviderType)
	if err != nil {
		return "", "", fmt.Errorf("创建DNS提供者失败: %v", err)
	}

	err = c.client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		return "", "", fmt.Errorf("设置DNS挑战提供者失败: %v", err)
	}

	// 创建证书请求
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	certificates, err := c.client.Certificate.Obtain(request)
	if err != nil {
		return "", "", fmt.Errorf("获取证书失败: %v", err)
	}

	return string(certificates.Certificate), string(certificates.PrivateKey), nil
}

// RenewCertificate 使用ACME Renewal Information (ARI)协议续期SSL证书
// 实现了RFC 8555定义的ARI协议，支持基于标准CertID的智能续期
// domain: 证书域名
// dnsProvider: DNS提供者信息结构体，包含ProviderType、APIKey和APISecret
// currentCert: 当前证书的PEM格式内容，用于ARI检查
// 返回证书内容、私钥和可能的错误信息
func (c *ACMEClient) RenewCertificate(domain string, dnsProvider CertificateRequestInfo, currentCert string) (string, string, error) {
	// 如果没有当前证书，直接执行标准申请
	if currentCert == "" {
		return c.RequestCertificate(domain, dnsProvider)
	}

	// 解析证书并生成ARI CertID
	cert, err := parseCertificate(currentCert)
	if err != nil {
		return "", "", fmt.Errorf("解析证书失败: %v", err)
	}

	ariCertID, err := certificate.MakeARICertID(cert)
	if err != nil {
		return "", "", fmt.Errorf("生成ARI CertID失败: %v", err)
	}

	// 检查是否需要续期
	shouldRenew := time.Until(cert.NotAfter) < 30*24*time.Hour
	if !shouldRenew {
		return "", "", fmt.Errorf("根据证书有效期检查，当前证书暂不需要续期")
	}

	// 设置DNS提供者
	factory := NewDNSProviderFactory(dnsProvider.APIKey, dnsProvider.APISecret, dnsProvider.APIToken)
	provider, err := factory.CreateProvider(dnsProvider.ProviderType)
	if err != nil {
		return "", "", fmt.Errorf("创建DNS提供者失败: %v", err)
	}

	err = c.client.Challenge.SetDNS01Provider(provider)
	if err != nil {
		return "", "", fmt.Errorf("设置DNS挑战提供者失败: %v", err)
	}

	// 使用ARI协议执行续期
	request := certificate.ObtainRequest{
		Domains:        []string{domain},
		Bundle:         true,
		ReplacesCertID: ariCertID, // 使用标准ARI CertID进行续期
	}

	certificates, err := c.client.Certificate.Obtain(request)
	if err != nil {
		return "", "", fmt.Errorf("使用ARI协议续期证书失败: %v", err)
	}

	return string(certificates.Certificate), string(certificates.PrivateKey), nil
}

// parseCertificate 解析PEM格式的证书
func parseCertificate(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("无效的PEM格式证书")
	}

	return x509.ParseCertificate(block.Bytes)
}
