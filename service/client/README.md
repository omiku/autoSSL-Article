# SSL客户端使用指南

## 概述

本目录提供了统一的SSL证书管理客户端，支持多个证书颁发机构（CA），包括Let's Encrypt和ZeroSSL。

## 目录结构

```
service/client/
├── client.go              # 统一的客户端接口和工厂
├── client_test.go         # 测试文件
├── acme_client.go         # ACME协议客户端实现
├── acme_account_manager.go # ACME账户管理器
├── dns_provider_factory.go # DNS提供商工厂
├── ratelimiter.go         # 速率限制器
└── ratelimiter_test.go   # 速率限制器测试
```

## 使用方法

### 1. 创建客户端工厂

```go
import "autoSSL/service/client"

// 创建工厂
factory := client.NewFactory(entClient)

// 获取支持的提供商
providers := factory.GetSupportedProviders()
// 返回: ["letsencrypt", "zerossl"]
```

### 2. 获取SSL客户端

```go
// 获取Let's Encrypt客户端
letsEncryptClient, err := factory.GetClient(ctx, userID, "user@example.com", client.ProviderLetsEncrypt)
if err != nil {
    // 处理错误
}

// 获取ZeroSSL客户端
zeroSSLClient, err := factory.GetClient(ctx, userID, "user@example.com", client.ProviderZeroSSL)
if err != nil {
    // 处理错误
}
```

### 3. 申请SSL证书

```go
// 配置DNS提供商信息
dnsInfo := client.CertificateRequestInfo{
    ProviderType: "cloudflare",
    APIKey:       "your-api-key",
    APISecret:    "your-api-secret",
    APIToken:     "your-api-token", // 可选
}

// 申请证书
cert, key, err := client.RequestCertificate("example.com", dnsInfo)
if err != nil {
    // 处理错误
}

// cert: 证书内容（PEM格式）
// key: 私钥内容（PEM格式）
```

### 4. 续期SSL证书

```go
// 续期证书（使用当前证书内容）
newCert, newKey, err := client.RenewCertificate("example.com", dnsInfo, currentCertContent)
if err != nil {
    // 处理错误
}
```

## 支持的证书提供商

| 提供商 | 常量值 | ACME目录URL |
|--------|--------|-------------|
| Let's Encrypt | `client.ProviderLetsEncrypt` | https://acme-v02.api.letsencrypt.org/directory |
| Let's Encrypt (Staging) | `"letsencrypt-staging"` | https://acme-staging-v02.api.letsencrypt.org/directory |
| ZeroSSL | `client.ProviderZeroSSL` | https://acme.zerossl.com/v2/DV90 |

## 支持的DNS提供商

通过 `dns_provider_factory.go` 支持以下DNS提供商：

- Cloudflare
- AliDNS (阿里云DNS)
- DNSPod (腾讯云DNS)
- 以及其他标准DNS提供商

## 错误处理

所有方法都会返回详细的错误信息，建议按以下方式处理：

```go
cert, key, err := client.RequestCertificate(domain, dnsInfo)
if err != nil {
    log.Printf("申请证书失败: %v", err)
    // 根据错误类型进行相应处理
}
```

## 数据库集成

系统使用Ent ORM管理ACME账户信息，自动处理：
- 账户创建和复用
- 私钥和注册信息管理
- 多提供商支持
- 用户关联

## 速率限制

内置速率限制器防止：
- 频繁的ACME账户创建
- 过度的API调用
- 资源滥用

## 向后兼容性

为了保持向后兼容，`GetOrCreateACMEAccount` 方法仍然可用，默认使用Let's Encrypt：

```go
// 旧代码（仍然可用）
client, err := acmeManager.GetOrCreateACMEAccount(ctx, userID, email)

// 新代码（推荐）
client, err := factory.GetClient(ctx, userID, email, client.ProviderLetsEncrypt)
```