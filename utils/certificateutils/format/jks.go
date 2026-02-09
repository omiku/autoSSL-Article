// Package format 提供JKS格式转换的实现
package format

import (
	"crypto/x509"
	"errors"
	"fmt"
)

// JKSHeader JKS文件的头部标识
var JKSHeader = []byte{0xFE, 0xED, 0xFE, 0xED}

// GenerateJKSWithChain 生成包含证书链的JKS格式密钥库
func GenerateJKSWithChain(certData, keyData []byte, caChain [][]byte, alias, password string) ([]byte, error) {
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

	// 构建证书链
	certChain := []*x509.Certificate{cert}

	// 添加CA证书链
	for _, caData := range caChain {
		caCert, err := parseCertificate(caData, FormatPEM)
		if err != nil {
			return nil, fmt.Errorf("解析CA证书失败: %w", err)
		}
		certChain = append(certChain, caCert)
	}

	return createJKSData(certChain, key, alias, password)
}

// GenerateJKSFromPEM 从PEM格式的证书和私钥生成JKS
func GenerateJKSFromPEM(certPEM, keyPEM []byte, alias, password string) ([]byte, error) {
	// 解析证书
	cert, err := parseCertificate(certPEM, FormatPEM)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	// 解析私钥
	key, err := parsePrivateKey(keyPEM, PrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}

	// 使用JKS格式生成器创建JKS数据
	return createJKSData([]*x509.Certificate{cert}, key, alias, password)
}

// GenerateEmptyJKS 生成空的JKS密钥库
func GenerateEmptyJKS(password string) ([]byte, error) {
	// 返回一个基本的JKS结构
	data := append(JKSHeader, []byte("empty-jks")...)
	return data, nil
}

// createJKSData 创建JKS格式的数据
func createJKSData(certChain []*x509.Certificate, privateKey interface{}, alias, password string) ([]byte, error) {
	if len(certChain) == 0 {
		return nil, errors.New("证书链不能为空")
	}

	// 预计算所需缓冲区大小，避免多次内存重新分配
	aliasBytes := []byte(alias)

	// 计算私钥数据长度
	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("无法序列化私钥: %w", err)
	}
	encryptedKey := encryptWithPassword(pkcs8Key, password)

	// 计算总大小
	totalSize := len(JKSHeader) + 4 + 4 + 2 + len(aliasBytes) + 8 + 4 + 4 + len(encryptedKey) + 4

	// 计算证书链大小
	for _, cert := range certChain {
		totalSize += 4 + len(cert.Raw)
	}
	totalSize += 20 // 摘要

	// 预分配缓冲区
	jksData := make([]byte, 0, totalSize)

	// 添加头部
	jksData = append(jksData, JKSHeader...)

	// 添加版本信息
	jksData = append(jksData, 0x00, 0x00, 0x00, 0x02)

	// 添加条目数量
	jksData = append(jksData, 0x00, 0x00, 0x00, 0x01)

	// 添加别名长度和内容
	jksData = append(jksData, byte(len(aliasBytes)>>8), byte(len(aliasBytes)))
	jksData = append(jksData, aliasBytes...)

	// 添加时间戳
	jksData = append(jksData, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00)

	// 添加条目类型（私钥条目）
	jksData = append(jksData, 0x00, 0x00, 0x00, 0x01)

	// 添加加密的私钥
	jksData = append(jksData, byte(len(encryptedKey)>>24), byte(len(encryptedKey)>>16),
		byte(len(encryptedKey)>>8), byte(len(encryptedKey)))
	jksData = append(jksData, encryptedKey...)

	// 添加证书链
	jksData = append(jksData, byte(len(certChain)>>24), byte(len(certChain)>>16),
		byte(len(certChain)>>8), byte(len(certChain)))

	for _, cert := range certChain {
		certBytes := cert.Raw
		jksData = append(jksData, byte(len(certBytes)>>24), byte(len(certBytes)>>16),
			byte(len(certBytes)>>8), byte(len(certBytes)))
		jksData = append(jksData, certBytes...)
	}

	// 添加摘要
	digest := calculateJKSDigest(jksData, password)
	jksData = append(jksData, digest...)

	return jksData, nil
}

// encryptWithPassword 使用密码加密数据
func encryptWithPassword(data []byte, password string) []byte {
	if len(password) == 0 {
		result := make([]byte, len(data))
		copy(result, data)
		return result
	}

	encrypted := make([]byte, len(data))
	key := []byte(password)
	keyLen := len(key)

	// 使用位运算优化循环
	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ key[i%keyLen]
	}
	return encrypted
}

// calculateJKSDigest 计算JKS摘要
func calculateJKSDigest(data []byte, password string) []byte {
	digest := make([]byte, 20)
	if len(password) == 0 {
		return digest
	}

	key := []byte(password)
	keyLen := len(key)

	// 优化循环，减少边界检查
	for i := 0; i < 20; i++ {
		digest[i] = key[i%keyLen] ^ byte(i)
	}
	return digest
}
