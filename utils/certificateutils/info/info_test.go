package info

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func TestParseCertificateFromPEM(t *testing.T) {
	// 生成测试证书
	certPEM := generateTestCertificatePEM(t)
	
	// 解析证书信息
	info, err := ParseCertificateFromPEM(certPEM)
	if err != nil {
		t.Fatalf("解析证书失败: %v", err)
	}
	
	// 验证解析结果
	if info.Subject == "" {
		t.Error("主题为空")
	}
	
	if info.Issuer == "" {
		t.Error("颁发者为空")
	}
	
	if info.SerialNumber == "" {
		t.Error("序列号为空")
	}
	
	if info.NotAfter.Before(info.NotBefore) {
		t.Error("过期时间在生效时间之前")
	}
	
	if info.DaysUntilExpiry <= 0 {
		t.Error("剩余天数计算错误")
	}
	
	if info.Status != "valid" {
		t.Errorf("证书状态应为valid，实际为: %s", info.Status)
	}
}

func TestParseCertificateFromDER(t *testing.T) {
	// 生成测试证书
	certDER := generateTestCertificateDER(t)
	
	// 解析证书信息
	info, err := ParseCertificateFromDER(certDER)
	if err != nil {
		t.Fatalf("解析证书失败: %v", err)
	}
	
	// 验证解析结果
	if info.Subject == "" {
		t.Error("主题为空")
	}
}

func TestValidateCertificate(t *testing.T) {
	// 生成有效证书
	certPEM := generateTestCertificatePEM(t)
	
	// 验证有效证书
	err := ValidateCertificate(certPEM)
	if err != nil {
		t.Errorf("验证有效证书失败: %v", err)
	}
	
	// 生成过期证书
	expiredCert := generateExpiredCertificatePEM(t)
	
	// 验证过期证书
	err = ValidateCertificate(expiredCert)
	if err == nil {
		t.Error("过期证书验证应失败")
	}
}

func TestGetCertificateExpiry(t *testing.T) {
	certPEM := generateTestCertificatePEM(t)
	
	expiry, days, err := GetCertificateExpiry(certPEM)
	if err != nil {
		t.Fatalf("获取过期信息失败: %v", err)
	}
	
	if expiry.IsZero() {
		t.Error("过期时间为零")
	}
	
	if days <= 0 {
		t.Error("剩余天数计算错误")
	}
}

func TestCheckCertificateDomain(t *testing.T) {
	certPEM := generateTestCertificatePEM(t)
	
	// 测试匹配域名
	matched, err := CheckCertificateDomain(certPEM, "test.example.com")
	if err != nil {
		t.Fatalf("检查域名失败: %v", err)
	}
	
	if !matched {
		t.Error("应匹配域名 test.example.com")
	}
	
	// 测试不匹配域名
	matched, err = CheckCertificateDomain(certPEM, "invalid.com")
	if err != nil {
		t.Fatalf("检查域名失败: %v", err)
	}
	
	if matched {
		t.Error("不应匹配域名 invalid.com")
	}
}

// generateTestCertificatePEM 生成测试PEM格式证书
func generateTestCertificatePEM(t *testing.T) []byte {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}
	
	template := &x509.Certificate{
		SerialNumber: big.NewInt(12345),
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
			CommonName:   "test.example.com",
		},
		DNSNames:    []string{"test.example.com", "www.example.com"},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(30 * 24 * time.Hour), // 30天后过期
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("创建证书失败: %v", err)
	}
	
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
}

// generateTestCertificateDER 生成测试DER格式证书
func generateTestCertificateDER(t *testing.T) []byte {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}
	
	template := &x509.Certificate{
		SerialNumber: big.NewInt(12345),
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
			CommonName:   "test.example.com",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(30 * 24 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("创建证书失败: %v", err)
	}
	
	return certDER
}

// generateExpiredCertificatePEM 生成过期证书用于测试
func generateExpiredCertificatePEM(t *testing.T) []byte {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}
	
	template := &x509.Certificate{
		SerialNumber: big.NewInt(12346),
		Subject: pkix.Name{
			Organization: []string{"Test Organization"},
			CommonName:   "expired.example.com",
		},
		NotBefore: time.Now().Add(-30 * 24 * time.Hour), // 30天前生效
		NotAfter:  time.Now().Add(-1 * time.Hour),        // 1小时前过期
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("创建证书失败: %v", err)
	}
	
	return pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
}