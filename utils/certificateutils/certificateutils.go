// Package certificateutils 提供证书相关的工具函数
package certificateutils

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"time"

	"autoSSL/utils/certificateutils/format"
	"autoSSL/utils/certificateutils/info"
)

// CertificateFormat 重导出格式转换相关类型
var (
	FormatPEM    = format.FormatPEM
	FormatDER    = format.FormatDER
	FormatPFX    = format.FormatPFX
	FormatPKCS12 = format.FormatPKCS12
	FormatJKS    = format.FormatJKS
)

// PrivateKeyFormat 重导出私钥格式相关类型
var (
	PrivateKeyPEM   = format.PrivateKeyPEM
	PrivateKeyDER   = format.PrivateKeyDER
	PrivateKeyPKCS8 = format.PrivateKeyPKCS8
	PrivateKeyPKCS1 = format.PrivateKeyPKCS1
)

// ConvertCertificate 转换证书格式
func ConvertCertificate(certData []byte, fromFormat, toFormat string) ([]byte, error) {
	return format.ConvertCertificate(certData, format.CertificateFormat(fromFormat), format.CertificateFormat(toFormat))
}

// GenerateJKS 生成JKS格式的密钥库
func GenerateJKS(certData, keyData []byte, alias, password string) ([]byte, error) {
	return format.GenerateJKS(certData, keyData, alias, password)
}

// ConvertPrivateKey 转换私钥格式
func ConvertPrivateKey(keyData []byte, fromFormat, toFormat string) ([]byte, error) {
	return format.ConvertPrivateKey(keyData, format.PrivateKeyFormat(fromFormat), format.PrivateKeyFormat(toFormat))
}

// ParseCertificate 解析证书信息
func ParseCertificate(certData []byte, isPEM bool) (*info.CertificateInfo, error) {
	if isPEM {
		return info.ParseCertificateFromPEM(certData)
	}
	return info.ParseCertificateFromDER(certData)
}

// ValidateCertificate 验证证书有效性
func ValidateCertificate(certData []byte, isPEM bool) error {
	if isPEM {
		return info.ValidateCertificate(certData)
	}
	cert, err := info.ParseCertificateFromDER(certData)
	if err != nil {
		return err
	}
	if cert.Status != "valid" {
		return fmt.Errorf("证书状态无效: %s", cert.Status)
	}
	return nil
}

// GetCertificateExpiry 获取证书过期信息
func GetCertificateExpiry(certData []byte, isPEM bool) (time.Time, int, error) {
	if isPEM {
		return info.GetCertificateExpiry(certData)
	}
	cert, err := info.ParseCertificateFromDER(certData)
	if err != nil {
		return time.Time{}, 0, err
	}
	return cert.NotAfter, cert.DaysUntilExpiry, nil
}

// CheckCertificateDomain 检查证书域名匹配
func CheckCertificateDomain(certData []byte, domain string, isPEM bool) (bool, error) {
	if isPEM {
		return info.CheckCertificateDomain(certData, domain)
	}
	_, err := info.ParseCertificateFromDER(certData)
	if err != nil {
		return false, err
	}
	return info.CheckCertificateDomain(certData, domain)
}

// GeneratePFX 生成PFX格式的证书包
func GeneratePFX(certData, keyData []byte, password string) ([]byte, error) {
	return format.GeneratePFX(certData, keyData, password)
}

// PackageCertificate 将证书打包为ZIP格式，支持多种格式导出
// 每个证书格式以单独文件夹存在，文件夹中包含该证书和对应的私钥
// 使用零拷贝方式在内存中构建压缩包，不生成临时文件
func PackageCertificate(certContent, privateKeyContent, password string) ([]byte, error) {
	// 创建内存缓冲区，避免磁盘IO
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// PEM格式文件夹
	if err := addFileToZip(zipWriter, "PEM/certificate.pem", []byte(certContent)); err != nil {
		return nil, fmt.Errorf("添加PEM证书文件失败: %v", err)
	}
	if err := addFileToZip(zipWriter, "PEM/private.key", []byte(privateKeyContent)); err != nil {
		return nil, fmt.Errorf("添加PEM私钥文件失败: %v", err)
	}

	// DER格式文件夹
	if derCert, err := ConvertCertificate([]byte(certContent), "pem", "der"); err == nil {
		addFileToZip(zipWriter, "DER/certificate.der", derCert)
		// DER格式私钥
		if derKey, err := ConvertPrivateKey([]byte(privateKeyContent), "pem", "der"); err == nil {
			addFileToZip(zipWriter, "DER/private.key", derKey)
		}
	}

	// PFX格式文件夹
	if pfxData, err := GeneratePFX([]byte(certContent), []byte(privateKeyContent), password); err == nil {
		addFileToZip(zipWriter, "PFX/certificate.pfx", pfxData)
		// 添加说明文件，因为PFX已包含私钥
		pfxInfo := fmt.Sprintf(`PFX格式说明
该文件包含证书和私钥，密码: %s
`, password)
		addFileToZip(zipWriter, "PFX/README.txt", []byte(pfxInfo))
	}

	// JKS格式文件夹
	if jksData, err := GenerateJKS([]byte(certContent), []byte(privateKeyContent), "certificate", password); err == nil {
		addFileToZip(zipWriter, "JKS/certificate.jks", jksData)
		// 添加说明文件
		jksInfo := fmt.Sprintf(`JKS格式说明
该文件包含证书和私钥，密码: %s
`, password)
		addFileToZip(zipWriter, "JKS/README.txt", []byte(jksInfo))
	}

	// PKCS8格式私钥文件夹
	if pkcs8Key, err := ConvertPrivateKey([]byte(privateKeyContent), "pem", "pkcs8"); err == nil {
		addFileToZip(zipWriter, "PKCS8/certificate.pem", []byte(certContent))
		addFileToZip(zipWriter, "PKCS8/private.pkcs8", pkcs8Key)
	}

	// 根目录添加总体说明文件
	infoContent := fmt.Sprintf(`证书包说明
生成时间: %s
包含格式: PEM, DER, PFX, JKS, PKCS8
PFX/JKS密码: %s

文件夹结构:
- PEM/: PEM格式的证书和私钥
- DER/: DER格式的证书和私钥
- PFX/: PFX格式的证书包（包含证书和私钥）
- JKS/: JKS格式的Java密钥库（包含证书和私钥）
- PKCS8/: PEM格式证书和PKCS8格式私钥
`,
		time.Now().Format("2006-01-02 15:04:05"), password)
	addFileToZip(zipWriter, "README.txt", []byte(infoContent))

	// 关闭zip写入器
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("关闭zip写入器失败: %v", err)
	}

	// 返回内存中的zip数据，实现零拷贝
	return buf.Bytes(), nil
}

// addFileToZip 添加文件到zip包的辅助函数
func addFileToZip(zipWriter *zip.Writer, filename string, content []byte) error {
	fileWriter, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}
	_, err = fileWriter.Write(content)
	return err
}

// 错误处理
var (
	ErrInvalidCertificate = errors.New("无效的证书")
	ErrInvalidPrivateKey  = errors.New("无效的私钥")
	ErrInvalidFormat      = errors.New("不支持的格式")
	ErrExpiredCertificate = errors.New("证书已过期")
	ErrNotYetValid        = errors.New("证书尚未生效")
)
