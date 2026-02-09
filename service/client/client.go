package client

import (
	"context"
	"fmt"

	"autoSSL/ent"
)

// Client 统一的SSL证书客户端接口
type Client interface {
	// RequestCertificate 申请SSL证书
	// domain: 证书域名
	// dnsProvider: DNS提供者配置信息
	// 返回证书内容、私钥和错误信息
	RequestCertificate(domain string, dnsProvider CertificateRequestInfo) (cert string, key string, err error)

	// RenewCertificate 续期SSL证书
	// domain: 证书域名
	// dnsProvider: DNS提供者配置信息
	// currentCert: 当前证书的PEM格式内容
	// 返回新证书内容、新私钥和错误信息
	RenewCertificate(domain string, dnsProvider CertificateRequestInfo, currentCert string) (cert string, key string, err error)
}

// Provider SSL证书提供商类型
type Provider string

const (
	ProviderLetsEncrypt Provider = "letsencrypt"
	ProviderZeroSSL     Provider = "zerossl"
)

// CertificateRequestInfo DNS提供者配置信息
type CertificateRequestInfo struct {
	ProviderType string // DNS提供者类型，如cloudflare、alidns等
	APIKey       string // API密钥
	APISecret    string // API密钥对应的Secret
	APIToken     string // API令牌（某些DNS提供商使用）
}

// Factory SSL客户端工厂接口
type Factory interface {
	// GetClient 获取指定提供商的SSL客户端
	GetClient(ctx context.Context, userID int, email string, provider Provider) (Client, error)
	
	// GetSupportedProviders 获取支持的SSL证书提供商列表
	GetSupportedProviders() []Provider
}

// NewFactory 创建SSL客户端工厂
func NewFactory(entClient *ent.Client) Factory {
	return &sslFactory{
		acmeManager: NewACMEAccountManager(entClient, NewRateLimiter()),
		entClient:   entClient,
	}
}

// sslFactory SSL客户端工厂实现
type sslFactory struct {
	acmeManager *ACMEAccountManager
	entClient   *ent.Client
}

// GetClient 获取指定提供商的SSL客户端
func (f *sslFactory) GetClient(ctx context.Context, userID int, email string, provider Provider) (Client, error) {
	switch provider {
	case ProviderLetsEncrypt, ProviderZeroSSL:
		return f.acmeManager.GetOrCreateAccountWithProvider(ctx, userID, email, string(provider))
	default:
		return nil, fmt.Errorf("不支持的SSL证书提供商: %s", provider)
	}
}

// GetSupportedProviders 获取支持的SSL证书提供商列表
func (f *sslFactory) GetSupportedProviders() []Provider {
	return []Provider{ProviderLetsEncrypt, ProviderZeroSSL}
}