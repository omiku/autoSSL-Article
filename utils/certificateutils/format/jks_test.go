package format

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

// generateTestCertAndKey 生成测试用的证书和私钥
func generateTestCertAndKey(t *testing.T) ([]byte, []byte) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024) // 使用1024位以减少测试时间
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}

	// 创建证书模板
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(24 * time.Hour),
		KeyUsage:  x509.KeyUsageDigitalSignature,
	}

	// 自签名证书
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("创建证书失败: %v", err)
	}

	// PEM编码证书
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	// PEM编码私钥
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return certPEM, keyPEM
}

// TestGenerateJKS 测试JKS生成
func TestGenerateJKS(t *testing.T) {
	certPEM, keyPEM := generateTestCertAndKey(t)

	alias := "test-alias"
	password := "test-password"

	// 生成JKS
	jksData, err := GenerateJKS(certPEM, keyPEM, alias, password)
	if err != nil {
		t.Fatalf("生成JKS失败: %v", err)
	}

	if len(jksData) == 0 {
		t.Fatal("生成的JKS数据为空")
	}

	// 验证JKS格式
	if len(jksData) < 8 {
		t.Fatal("JKS数据长度不足")
	}

	// 检查头部
	if len(jksData) < 4 {
		t.Fatal("JKS数据太短")
	}
	expectedHeader := []byte{0xFE, 0xED, 0xFE, 0xED}
	if string(jksData[:4]) != string(expectedHeader) {
		t.Errorf("JKS头部不正确")
	}
}

// TestGenerateJKSWithChain 测试带证书链的JKS生成
func TestGenerateJKSWithChain(t *testing.T) {
	certPEM, keyPEM := generateTestCertAndKey(t)

	alias := "test-alias"
	password := "test-password"

	// 生成JKS（不带CA链）
	jksData, err := GenerateJKSWithChain(certPEM, keyPEM, nil, alias, password)
	if err != nil {
		t.Fatalf("生成带证书链的JKS失败: %v", err)
	}

	if len(jksData) < 4 {
		t.Fatal("生成的JKS数据太短")
	}
}

// TestGenerateEmptyJKS 测试空JKS生成
func TestGenerateEmptyJKS(t *testing.T) {
	password := "test-password"

	jksData, err := GenerateEmptyJKS(password)
	if err != nil {
		t.Fatalf("生成空JKS失败: %v", err)
	}

	if len(jksData) == 0 {
		t.Fatal("生成的空JKS数据为空")
	}
}

// TestGenerateJKSFromPEM 测试从PEM生成JKS
func TestGenerateJKSFromPEM(t *testing.T) {
	certPEM, keyPEM := generateTestCertAndKey(t)

	alias := "test-alias"
	password := "test-password"

	jksData, err := GenerateJKSFromPEM(certPEM, keyPEM, alias, password)
	if err != nil {
		t.Fatalf("从PEM生成JKS失败: %v", err)
	}

	if len(jksData) == 0 {
		t.Fatal("生成的JKS数据为空")
	}
}

// TestGenerateJKSInvalidInput 测试无效输入
func TestGenerateJKSInvalidInput(t *testing.T) {
	certPEM := []byte("invalid-cert")
	keyPEM := []byte("invalid-key")

	alias := "test-alias"
	password := "test-password"

	_, err := GenerateJKS(certPEM, keyPEM, alias, password)
	if err == nil {
		t.Fatal("应该返回错误，但没有返回")
	}
}

// BenchmarkGenerateJKS 基准测试已移除以避免超时