package kuocai_cdn

import (
	"autoSSL/logger"
	"autoSSL/utils/deployment/providers"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type SSLDeployerProviderConfig struct {
	// 括彩云 账号。
	Account string `json:"account"`
	// 括彩云 密码。
	Password string `json:"password"`
	// 加速域名列表（支持泛域名）。
	Domains string `json:"domains"`
}

type SSLDeployerProvider struct {
	config *SSLDeployerProviderConfig
	logger *logger.Logger
	client *KuocaiClient
}

var _ providers.DeployProvider = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}
	client := NewKuocaiClient()
	if err := client.Login(config.Account, config.Password); err != nil {
		return nil, err
	}
	return &SSLDeployerProvider{
		config: config,
		logger: logger.GetLogger(),
		client: client,
	}, nil
}

func (s *SSLDeployerProvider) Name() string {
	return "kuocai-cdn"
}

func (s *SSLDeployerProvider) DeployCertificate(ctx context.Context, certContent, keyContent string, workflowConfig map[string]interface{}) error {
	s.logger.Info("开始部署证书到括彩云CDN", zap.String("domain", s.config.Domains))

	// 获取域名ID
	domainID, err := s.getDomainIDWithClient()
	if err != nil {
		return fmt.Errorf("获取域名ID失败: %w", err)
	}

	// 配置HTTPS证书
	err = s.configureHTTPSCertificate(domainID, certContent, keyContent)
	if err != nil {
		return fmt.Errorf("配置HTTPS证书失败: %w", err)
	}

	s.logger.Info("证书部署成功，正在生效中", zap.String("domain", s.config.Domains))
	return nil
}



// getDomainIDWithClient 获取加速域名id
func (s *SSLDeployerProvider) getDomainIDWithClient() (string, error) {
	requestBody := fmt.Sprintf(`{"draw":3,"columns":[],"start":0,"length":10,"search":{"value":"%s","regex":false}}`, s.config.Domains)
	resp, err := s.client.Post("/CdnDomain/queryForDatatables", requestBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var listDomainsResp ListDomainsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listDomainsResp); err != nil {
		return "", err
	}

	if listDomainsResp.Code != "SUCCESS" {
		return "", fmt.Errorf("list domains failed: %s", listDomainsResp.Message)
	}

	// 从解析后的结构体数组中筛选匹配域名的id
	for _, d := range listDomainsResp.Data.Data {
		if d.DomainName == s.config.Domains {
			return d.Id, nil
		}
	}
	return "", fmt.Errorf("domain %s not found", s.config.Domains)
}

// configureHTTPSCertificate 根据API格式配置HTTPS证书
func (s *SSLDeployerProvider) configureHTTPSCertificate(domainID, certContent, keyContent string) error {
	// certName为前缀+unixTime
	certName := fmt.Sprintf("autoSSL-%d", time.Now().Unix())
	requestBody := fmt.Sprintf(`{"doMainId":"%s","https":{"https_status":"on","certificate_source":"0","certificate_name":"%s","certificate_value":"%s","private_key":"%s"}}`,
		domainID, certName, certContent, keyContent)

	// 使用Post发送raw字符串
	resp, err := s.client.Post("/CdnDomainHttps/httpsConfiguration", requestBody)
	if err != nil {
		return fmt.Errorf("发送配置请求失败: %w", err)
	}
	defer resp.Body.Close()

	var respBody struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if respBody.Code != "SUCCESS" || !respBody.Success {
		return fmt.Errorf("配置HTTPS证书失败: %s", respBody.Message)
	}

	s.logger.Info("HTTPS证书配置成功",
		zap.String("domainID", domainID),
		zap.String("message", respBody.Message),
	)
	return nil
}
