package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestProvider 测试提供商常量
type TestProvider struct {
	name string
}

func TestProviderConstants(t *testing.T) {
	assert.Equal(t, Provider("letsencrypt"), ProviderLetsEncrypt)
	assert.Equal(t, Provider("zerossl"), ProviderZeroSSL)
}

func TestGetCADirectoryURL(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		want     string
	}{
		{"Let's Encrypt", "letsencrypt", "https://acme-v02.api.letsencrypt.org/directory"},
		{"Let's Encrypt Staging", "letsencrypt-staging", "https://acme-staging-v02.api.letsencrypt.org/directory"},
		{"ZeroSSL", "zerossl", "https://acme.zerossl.com/v2/DV90"},
		{"Unknown", "unknown", "https://acme-v02.api.letsencrypt.org/directory"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCADirectoryURL(tt.provider)
			assert.Equal(t, tt.want, got)
		})
	}
}

// MockClient 用于测试的模拟客户端
type MockClient struct {
	RequestCertificateFunc func(domain string, dnsProvider CertificateRequestInfo) (string, string, error)
	RenewCertificateFunc   func(domain string, dnsProvider CertificateRequestInfo, currentCert string) (string, string, error)
}

func (m *MockClient) RequestCertificate(domain string, dnsProvider CertificateRequestInfo) (string, string, error) {
	if m.RequestCertificateFunc != nil {
		return m.RequestCertificateFunc(domain, dnsProvider)
	}
	return "mock-cert", "mock-key", nil
}

func (m *MockClient) RenewCertificate(domain string, dnsProvider CertificateRequestInfo, currentCert string) (string, string, error) {
	if m.RenewCertificateFunc != nil {
		return m.RenewCertificateFunc(domain, dnsProvider, currentCert)
	}
	return "mock-cert-renewed", "mock-key-renewed", nil
}

func TestCertificateRequestInfo(t *testing.T) {
	info := CertificateRequestInfo{
		ProviderType: "cloudflare",
		APIKey:       "test-key",
		APISecret:    "test-secret",
		APIToken:     "test-token",
	}

	assert.Equal(t, "cloudflare", info.ProviderType)
	assert.Equal(t, "test-key", info.APIKey)
	assert.Equal(t, "test-secret", info.APISecret)
	assert.Equal(t, "test-token", info.APIToken)
}