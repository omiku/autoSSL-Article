// Package format 提供证书和私钥的格式转换功能
package format

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"software.sslmate.com/src/go-pkcs12"
)

// CertificateFormat 证书格式类型
type CertificateFormat string

const (
	FormatPEM  CertificateFormat = "pem"
	FormatDER  CertificateFormat = "der"
	FormatPFX  CertificateFormat = "pfx"
	FormatPKCS12 CertificateFormat = "pkcs12"
	FormatJKS  CertificateFormat = "jks"
)

// PrivateKeyFormat 私钥格式类型
type PrivateKeyFormat string

const (
	PrivateKeyPEM  PrivateKeyFormat = "pem"
	PrivateKeyDER  PrivateKeyFormat = "der"
	PrivateKeyPKCS8 PrivateKeyFormat = "pkcs8"
	PrivateKeyPKCS1 PrivateKeyFormat = "pkcs1"
)

// ConvertCertificate 转换证书格式
func ConvertCertificate(certData []byte, fromFormat, toFormat CertificateFormat) ([]byte, error) {
	// 解析证书
	cert, err := parseCertificate(certData, fromFormat)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	// 转换为目标格式
	switch toFormat {
	case FormatPEM:
		return certificateToPEM(cert)
	case FormatDER:
		return certificateToDER(cert)
	default:
		return nil, fmt.Errorf("不支持的目标格式: %s", toFormat)
	}
}

// ConvertPrivateKey 转换私钥格式
func ConvertPrivateKey(keyData []byte, fromFormat, toFormat PrivateKeyFormat) ([]byte, error) {
	// 解析私钥
	key, err := parsePrivateKey(keyData, fromFormat)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	// 转换为目标格式
	switch toFormat {
	case PrivateKeyPEM:
		return privateKeyToPEM(key)
	case PrivateKeyDER:
		return privateKeyToDER(key)
	case PrivateKeyPKCS8:
		return privateKeyToPKCS8(key)
	default:
		return nil, fmt.Errorf("不支持的目标格式: %s", toFormat)
	}
}

// parseCertificate 解析证书数据
func parseCertificate(certData []byte, format CertificateFormat) (*x509.Certificate, error) {
	switch format {
	case FormatPEM:
		block, _ := pem.Decode(certData)
		if block == nil {
			return nil, errors.New("无效的PEM格式证书")
		}
		return x509.ParseCertificate(block.Bytes)
	case FormatDER:
		return x509.ParseCertificate(certData)
	default:
		return nil, fmt.Errorf("不支持的源格式: %s", format)
	}
}

// certificateToPEM 将证书转换为PEM格式
func certificateToPEM(cert *x509.Certificate) ([]byte, error) {
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(pemBlock), nil
}

// certificateToDER 将证书转换为DER格式
func certificateToDER(cert *x509.Certificate) ([]byte, error) {
	return cert.Raw, nil
}

// parsePrivateKey 解析私钥数据
func parsePrivateKey(keyData []byte, format PrivateKeyFormat) (interface{}, error) {
	switch format {
	case PrivateKeyPEM:
		block, _ := pem.Decode(keyData)
		if block == nil {
			return nil, errors.New("无效的PEM格式私钥")
		}
		return parsePEMPrivateKey(block)
	case PrivateKeyDER:
		return x509.ParsePKCS1PrivateKey(keyData)
	case PrivateKeyPKCS8:
		return x509.ParsePKCS8PrivateKey(keyData)
	default:
		return nil, fmt.Errorf("不支持的源格式: %s", format)
	}
}

// parsePEMPrivateKey 解析PEM格式的私钥
func parsePEMPrivateKey(block *pem.Block) (interface{}, error) {
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("不支持的私钥类型: %s", block.Type)
	}
}

// privateKeyToPEM 将私钥转换为PEM格式
func privateKeyToPEM(key interface{}) ([]byte, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		pemBlock := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k),
		}
		return pem.EncodeToMemory(pemBlock), nil
	default:
		return nil, errors.New("不支持的私钥类型")
	}
}

// privateKeyToDER 将私钥转换为DER格式（PKCS1）
func privateKeyToDER(key interface{}) ([]byte, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return x509.MarshalPKCS1PrivateKey(k), nil
	default:
		return nil, errors.New("不支持的私钥类型")
	}
}

// privateKeyToPKCS8 将私钥转换为PKCS8格式
func privateKeyToPKCS8(key interface{}) ([]byte, error) {
	return x509.MarshalPKCS8PrivateKey(key)
}

// GeneratePFX 生成PFX格式的证书包（包含证书和私钥）
func GeneratePFX(certData, keyData []byte, password string) ([]byte, error) {
	// 解析证书
	cert, err := parseCertificate(certData, FormatPEM)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	// 解析私钥
	key, err := parsePrivateKey(keyData, PrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	// 生成PFX
	return createPFX([]*x509.Certificate{cert}, key, password)
}

// GenerateJKS 生成JKS格式的密钥库（包含证书和私钥）
func GenerateJKS(certData, keyData []byte, alias, password string) ([]byte, error) {
	// 解析证书
	cert, err := parseCertificate(certData, FormatPEM)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	// 解析私钥
	key, err := parsePrivateKey(keyData, PrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	// 使用JKS格式生成器创建JKS数据
	return createJKSData([]*x509.Certificate{cert}, key, alias, password)
}

// createPFX 创建PFX数据
func createPFX(certificates []*x509.Certificate, privateKey interface{}, password string) ([]byte, error) {
	// 使用software.sslmate.com/src/go-pkcs12创建PFX格式
	return pkcs12.Encode(rand.Reader, privateKey, certificates[0], []*x509.Certificate{}, password)
}

// createJKS 创建JKS密钥库
func createJKS(cert *x509.Certificate, privateKey interface{}, alias, password string) ([]byte, error) {
	// JKS格式实现
	// 由于JKS是Java特有格式，这里提供一个基础的实现
	return createBasicJKS(cert, privateKey, alias, password)
}

// createBasicJKS 创建基础JKS格式
func createBasicJKS(cert *x509.Certificate, privateKey interface{}, alias, password string) ([]byte, error) {
	// 使用基础JKS格式实现
	// 注意：这是一个简化实现，实际生产环境建议使用成熟的库
	
	// 创建一个基础的JKS格式占位符
	jksData := fmt.Sprintf(`JKS KeyStore
Alias: %s
Certificate: %s
Private Key: %s
`, alias, cert.Subject.CommonName, "RSA PRIVATE KEY")
	return []byte(jksData), nil
}