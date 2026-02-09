package format

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"runtime"
	"testing"
	"time"
)

func TestConvertCertificate(t *testing.T) {
	// 生成测试证书
	cert, _ := generateTestCertificate(t)
	
	// PEM转DER
	derData, err := certificateToDER(cert)
	if err != nil {
		t.Fatalf("PEM转DER失败: %v", err)
	}
	
	// DER转回PEM
	pemData, err := certificateToPEM(cert)
	if err != nil {
		t.Fatalf("DER转PEM失败: %v", err)
	}
	
	// 验证转换结果
	if len(derData) == 0 {
		t.Error("DER格式数据为空")
	}
	
	if len(pemData) == 0 {
		t.Error("PEM格式数据为空")
	}
	
	// 验证转换的正确性
	parsedCert, err := x509.ParseCertificate(derData)
	if err != nil {
		t.Fatalf("解析DER证书失败: %v", err)
	}
	
	if parsedCert.SerialNumber.String() != cert.SerialNumber.String() {
		t.Error("DER转换后的证书序列号不匹配")
	}
}

func TestConvertPrivateKey(t *testing.T) {
	// 生成测试私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}
	
	// PEM格式
	pemData, err := privateKeyToPEM(privateKey)
	if err != nil {
		t.Fatalf("私钥转PEM失败: %v", err)
	}
	
	// DER格式
	derData, err := privateKeyToDER(privateKey)
	if err != nil {
		t.Fatalf("私钥转DER失败: %v", err)
	}
	
	// PKCS8格式
	pkcs8Data, err := privateKeyToPKCS8(privateKey)
	if err != nil {
		t.Fatalf("私钥转PKCS8失败: %v", err)
	}
	
	// 验证转换结果
	if len(pemData) == 0 {
		t.Error("PEM格式私钥为空")
	}
	
	if len(derData) == 0 {
		t.Error("DER格式私钥为空")
	}
	
	if len(pkcs8Data) == 0 {
		t.Error("PKCS8格式私钥为空")
	}
	
	// 验证转换的正确性
	parsedKey, err := x509.ParsePKCS1PrivateKey(derData)
	if err != nil {
		t.Fatalf("解析DER私钥失败: %v", err)
	}
	
	if parsedKey.PublicKey.N.Cmp(privateKey.PublicKey.N) != 0 {
		t.Error("DER转换后的私钥公钥不匹配")
	}
}

func TestParseCertificate(t *testing.T) {
	// 生成测试证书
	cert, certPEM := generateTestCertificate(t)
	
	// 解析PEM格式
	parsedCert, err := parseCertificate(certPEM, FormatPEM)
	if err != nil {
		t.Fatalf("解析PEM证书失败: %v", err)
	}
	
	if parsedCert.SerialNumber.String() != cert.SerialNumber.String() {
		t.Error("证书序列号不匹配")
	}
}

func TestParsePrivateKey(t *testing.T) {
	// 生成测试私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}
	
	// 转换为PEM
	pemData, err := privateKeyToPEM(privateKey)
	if err != nil {
		t.Fatalf("私钥转PEM失败: %v", err)
	}
	
	// 解析PEM私钥
	parsedKey, err := parsePrivateKey(pemData, PrivateKeyPEM)
	if err != nil {
		t.Fatalf("解析PEM私钥失败: %v", err)
	}
	
	if parsedKey == nil {
		t.Error("解析后的私钥为nil")
	}
}

// generateTestCertificate 生成测试证书
func generateTestCertificate(t *testing.T) (*x509.Certificate, []byte) {
	// 强制垃圾回收，减少内存压力
	runtime.GC()
	
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024) // 使用较小的密钥减少内存使用
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}
	
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			CommonName:   "test.example.com",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(1 * time.Hour), // 减少证书有效期
		KeyUsage:  x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("创建证书失败: %v", err)
	}
	
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})
	
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("解析证书失败: %v", err)
	}
	
	return cert, certPEM
}