// Package info 提供证书信息读取和解析功能
package info

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"
)

// CertificateInfo 证书信息结构体
type CertificateInfo struct {
	SerialNumber       string    `json:"serial_number"`
	Subject            string    `json:"subject"`
	Issuer             string    `json:"issuer"`
	NotBefore          time.Time `json:"not_before"`
	NotAfter           time.Time `json:"not_after"`
	IsCA               bool      `json:"is_ca"`
	Version            int       `json:"version"`
	SignatureAlgorithm string    `json:"signature_algorithm"`
	PublicKeyAlgorithm string    `json:"public_key_algorithm"`
	KeyUsage           []string  `json:"key_usage"`
	ExtendedKeyUsage   []string  `json:"extended_key_usage"`
	DNSNames           []string  `json:"dns_names"`
	IPAddresses        []string  `json:"ip_addresses"`
	EmailAddresses     []string  `json:"email_addresses"`
	URIs               []string  `json:"uris"`
	Thumbprint         string    `json:"thumbprint"`
	SubjectKeyId       string    `json:"subject_key_id"`
	AuthorityKeyId     string    `json:"authority_key_id"`
	DaysUntilExpiry    int       `json:"days_until_expiry"`
	Status             string    `json:"status"`
}

// ParseCertificateFromPEM 从PEM格式解析证书信息
func ParseCertificateFromPEM(certPEM []byte) (*CertificateInfo, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, errors.New("无效的PEM格式")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	return parseCertificateInfo(cert), nil
}

// ParseCertificateFromDER 从DER格式解析证书信息
func ParseCertificateFromDER(certDER []byte) (*CertificateInfo, error) {
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	return parseCertificateInfo(cert), nil
}

// parseCertificateInfo 解析证书详细信息
func parseCertificateInfo(cert *x509.Certificate) *CertificateInfo {
	info := &CertificateInfo{
		SerialNumber:       cert.SerialNumber.String(),
		Subject:            cert.Subject.String(),
		Issuer:             cert.Issuer.String(),
		NotBefore:          cert.NotBefore,
		NotAfter:           cert.NotAfter,
		IsCA:               cert.IsCA,
		Version:            cert.Version,
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		DNSNames:           cert.DNSNames,
		EmailAddresses:     cert.EmailAddresses,
		Thumbprint:         fmt.Sprintf("%x", cert.Signature),
		DaysUntilExpiry:    int(time.Until(cert.NotAfter).Hours() / 24),
	}

	// 设置证书状态
	now := time.Now()
	if now.Before(cert.NotBefore) {
		info.Status = "not_yet_valid"
	} else if now.After(cert.NotAfter) {
		info.Status = "expired"
	} else {
		info.Status = "valid"
	}

	// 解析IP地址
	for _, ip := range cert.IPAddresses {
		info.IPAddresses = append(info.IPAddresses, ip.String())
	}

	// 解析URI
	for _, uri := range cert.URIs {
		info.URIs = append(info.URIs, uri.String())
	}

	// 解析KeyUsage
	info.KeyUsage = parseKeyUsage(cert.KeyUsage)

	// 解析ExtendedKeyUsage
	info.ExtendedKeyUsage = parseExtendedKeyUsage(cert.ExtKeyUsage)

	// 解析Key IDs
	if len(cert.SubjectKeyId) > 0 {
		info.SubjectKeyId = fmt.Sprintf("%x", cert.SubjectKeyId)
	}
	if len(cert.AuthorityKeyId) > 0 {
		info.AuthorityKeyId = fmt.Sprintf("%x", cert.AuthorityKeyId)
	}

	return info
}

// parseKeyUsage 解析KeyUsage
func parseKeyUsage(usage x509.KeyUsage) []string {
	var usages []string
	if usage&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "digital_signature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "content_commitment")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "key_encipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "data_encipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "key_agreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "cert_sign")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "crl_sign")
	}
	if usage&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "encipher_only")
	}
	if usage&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "decipher_only")
	}
	return usages
}

// parseExtendedKeyUsage 解析ExtendedKeyUsage
func parseExtendedKeyUsage(extKeyUsage []x509.ExtKeyUsage) []string {
	var usages []string
	for _, usage := range extKeyUsage {
		switch usage {
		case x509.ExtKeyUsageAny:
			usages = append(usages, "any")
		case x509.ExtKeyUsageServerAuth:
			usages = append(usages, "server_auth")
		case x509.ExtKeyUsageClientAuth:
			usages = append(usages, "client_auth")
		case x509.ExtKeyUsageCodeSigning:
			usages = append(usages, "code_signing")
		case x509.ExtKeyUsageEmailProtection:
			usages = append(usages, "email_protection")
		case x509.ExtKeyUsageIPSECEndSystem:
			usages = append(usages, "ipsec_end_system")
		case x509.ExtKeyUsageIPSECTunnel:
			usages = append(usages, "ipsec_tunnel")
		case x509.ExtKeyUsageIPSECUser:
			usages = append(usages, "ipsec_user")
		case x509.ExtKeyUsageTimeStamping:
			usages = append(usages, "time_stamping")
		case x509.ExtKeyUsageOCSPSigning:
			usages = append(usages, "ocsp_signing")
		case x509.ExtKeyUsageMicrosoftServerGatedCrypto:
			usages = append(usages, "microsoft_server_gated_crypto")
		case x509.ExtKeyUsageNetscapeServerGatedCrypto:
			usages = append(usages, "netscape_server_gated_crypto")
		}
	}
	return usages
}

// ValidateCertificate 验证证书有效性
func ValidateCertificate(certPEM []byte) error {
	info, err := ParseCertificateFromPEM(certPEM)
	if err != nil {
		return err
	}

	if info.Status != "valid" {
		return fmt.Errorf("证书状态无效: %s", info.Status)
	}

	return nil
}

// GetCertificateExpiry 获取证书过期时间
func GetCertificateExpiry(certPEM []byte) (time.Time, int, error) {
	info, err := ParseCertificateFromPEM(certPEM)
	if err != nil {
		return time.Time{}, 0, err
	}

	return info.NotAfter, info.DaysUntilExpiry, nil
}

// CheckCertificateDomain 检查证书是否匹配指定域名
func CheckCertificateDomain(certPEM []byte, domain string) (bool, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return false, errors.New("无效的PEM格式")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("解析证书失败: %w", err)
	}

	// 检查DNSNames
	for _, dns := range cert.DNSNames {
		if dns == domain {
			return true, nil
		}
	}

	// 检查CommonName
	if cert.Subject.CommonName == domain {
		return true, nil
	}

	// 通配符检查（简化版）
	if len(cert.DNSNames) > 0 {
		for _, dns := range cert.DNSNames {
			if matchWildcard(dns, domain) {
				return true, nil
			}
		}
	}

	return false, nil
}

// matchWildcard 通配符匹配检查
func matchWildcard(pattern, domain string) bool {
	if pattern == domain {
		return true
	}
	
	if len(pattern) > 2 && pattern[0] == '*' && pattern[1] == '.' {
		if len(domain) > len(pattern[2:]) {
			return domain[len(domain)-len(pattern[2:]):] == pattern[2:]
		}
	}
	
	return false
}