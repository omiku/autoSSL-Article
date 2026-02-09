# Certificate Utils

证书工具包，提供证书格式转换和证书信息读取功能。

## 功能特性

- **证书格式转换**: 支持PEM、DER、PFX、PKCS12、JKS等格式之间的转换
- **私钥格式转换**: 支持PEM、DER、PKCS8、PKCS1等格式之间的转换
- **证书信息读取**: 读取证书的详细信息，包括有效期、颁发者、主题、SAN等
- **证书验证**: 验证证书的有效性和域名匹配

## 目录结构

```
certificateutils/
├── format/          # 证书格式转换
├── info/           # 证书信息读取
├── certificateutils.go  # 统一入口
└── README.md       # 使用文档
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 使用示例

#### 证书格式转换

```go
package main

import (
	"e:\GolandProjects\autoSSL\utils\certificateutils"
	"fmt"
)

func main() {
	// PEM转DER
	pemCert := []byte(`-----BEGIN CERTIFICATE-----
...证书内容...
-----END CERTIFICATE-----`)
	
	derCert, err := certificateutils.ConvertCertificate(pemCert, "pem", "der")
	if err != nil {
		fmt.Printf("转换失败: %v\n", err)
		return
	}
	
	fmt.Printf("DER格式证书长度: %d\n", len(derCert))
}
```

#### 读取证书信息

```go
package main

import (
	"e:\GolandProjects\autoSSL\utils\certificateutils"
	"fmt"
)

func main() {
	certPEM := []byte(`-----BEGIN CERTIFICATE-----
...证书内容...
-----END CERTIFICATE-----`)
	
	info, err := certificateutils.ParseCertificate(certPEM, true)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
		return
	}
	
	fmt.Printf("证书主题: %s\n", info.Subject)
	fmt.Printf("颁发者: %s\n", info.Issuer)
	fmt.Printf("有效期: %s 至 %s\n", info.NotBefore, info.NotAfter)
	fmt.Printf("剩余天数: %d\n", info.DaysUntilExpiry)
	fmt.Printf("DNS名称: %v\n", info.DNSNames)
}
```

#### 验证证书

```go
package main

import (
	"e:\\GolandProjects\\autoSSL\\utils\\certificateutils"
	"fmt"
)

func main() {
	certPEM := []byte(`-----BEGIN CERTIFICATE-----
...证书内容...
-----END CERTIFICATE-----`)
	
	// 验证证书有效性
	err := certificateutils.ValidateCertificate(certPEM, true)
	if err != nil {
		fmt.Printf("证书无效: %v\n", err)
		return
	}
	
	// 检查域名匹配
	matched, err := certificateutils.CheckCertificateDomain(certPEM, "example.com", true)
	if err != nil {
		fmt.Printf("检查失败: %v\n", err)
		return
	}
	
	fmt.Printf("域名匹配: %v\n", matched)
}
```

#### 私钥格式转换

```go
package main

import (
    "io/ioutil"
    "log"
    
    "utils/certificateutils"
)

func main() {
    // 读取PEM格式的私钥
    keyData, err := ioutil.ReadFile("private.pem")
    if err != nil {
        log.Fatal(err)
    }
    
    // 转换为PKCS8格式
    pkcs8Key, err := certificateutils.ConvertPrivateKey(keyData, "pem", "pkcs8")
    if err != nil {
        log.Fatal(err)
    }
    
    // 保存为PKCS8格式
    if err := ioutil.WriteFile("private.pkcs8", pkcs8Key, 0600); err != nil {
        log.Fatal(err)
    }
    
    log.Println("私钥转换完成")
}
```

#### 生成JKS格式密钥库

```go
package main

import (
    "io/ioutil"
    "log"
    
    "utils/certificateutils"
)

func main() {
    // 读取证书和私钥
    certData, err := ioutil.ReadFile("certificate.pem")
    if err != nil {
        log.Fatal(err)
    }
    
    keyData, err := ioutil.ReadFile("private.pem")
    if err != nil {
        log.Fatal(err)
    }
    
    // 生成JKS格式密钥库
    jksData, err := certificateutils.GenerateJKS(certData, keyData, "myalias", "changeit")
    if err != nil {
        log.Fatal(err)
    }
    
    // 保存JKS文件
    if err := ioutil.WriteFile("keystore.jks", jksData, 0600); err != nil {
        log.Fatal(err)
    }
    
    log.Println("JKS密钥库生成完成")
}
```

## API文档

### 函数列表

#### ConvertCertificate
转换证书格式
```go
func ConvertCertificate(certData []byte, fromFormat, toFormat string) ([]byte, error)
```

**参数:**
- `certData`: 证书数据
- `fromFormat`: 源格式 ("pem", "der", "pfx", "pkcs12", "jks")
- `toFormat`: 目标格式 ("pem", "der")

#### ConvertPrivateKey
转换私钥格式
```go
func ConvertPrivateKey(keyData []byte, fromFormat, toFormat string) ([]byte, error)
```

**参数:**
- `keyData`: 私钥数据
- `fromFormat`: 源格式 ("pem", "der", "pkcs8", "pkcs1")
- `toFormat`: 目标格式 ("pem", "der", "pkcs8", "pkcs1")

#### ParseCertificate
解析证书信息
```go
func ParseCertificate(certData []byte, isPEM bool) (*CertificateInfo, error)
```

**参数:**
- `certData`: 证书数据
- `isPEM`: 是否为PEM格式

#### ValidateCertificate
验证证书有效性
```go
func ValidateCertificate(certData []byte, isPEM bool) error
```

#### GetCertificateExpiry
获取证书过期信息
```go
func GetCertificateExpiry(certData []byte, isPEM bool) (time.Time, int, error)
```

#### CheckCertificateDomain
检查证书域名匹配
```go
func CheckCertificateDomain(certData []byte, domain string, isPEM bool) (bool, error)
```

### CertificateInfo结构体

```go
type CertificateInfo struct {
	SerialNumber       string    // 序列号
	Subject            string    // 主题
	Issuer             string    // 颁发者
	NotBefore          time.Time // 生效时间
	NotAfter           time.Time // 过期时间
	IsCA               bool      // 是否为CA证书
	Version            int       // 版本
	SignatureAlgorithm string    // 签名算法
	PublicKeyAlgorithm string    // 公钥算法
	KeyUsage           []string  // 密钥用途
	ExtendedKeyUsage   []string  // 扩展密钥用途
	DNSNames           []string  // DNS名称
	IPAddresses        []string  // IP地址
	EmailAddresses     []string  // 邮箱地址
	URIs               []string  // URI
	Thumbprint         string    // 指纹
	SubjectKeyId       string    // 主题密钥ID
	AuthorityKeyId     string    // 颁发机构密钥ID
	DaysUntilExpiry    int       // 剩余天数
	Status             string    // 状态 (valid/expired/not_yet_valid)
}
```

## 注意事项

1. **PFX格式支持**: 当前PFX格式转换需要额外的`golang.org/x/crypto/pkcs12`包支持
2. **错误处理**: 所有函数都返回错误，使用时需要进行错误检查
3. **性能**: 大证书文件处理时注意内存使用
4. **安全性**: 处理私钥时注意安全，避免日志泄露

## 扩展建议

- 添加证书链验证功能
- 支持更多证书格式（如JKS）
- 添加证书吊销检查（CRL/OCSP）
- 支持批量证书处理