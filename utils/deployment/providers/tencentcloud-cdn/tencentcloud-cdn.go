package tencentcloudcdn

import (
	"autoSSL/logger"
	"autoSSL/utils/deployment/providers"
	"autoSSL/utils/deployment/providers/tencentcloudssl"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strings"

	tccdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
)

type SSLDeployerProviderConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 加速域名列表（支持泛域名）。
	Domains string `json:"domains"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *logger.Logger
	sdkClient  *tccdn.Client
	sslClient  *tcssl.Client
	sslManager tencentcloudssl.SSLManagerProviderInterface
}

var _ providers.DeployProvider = (*SSLDeployerProvider)(nil)

func NewSSLDeployerProvider(config *SSLDeployerProviderConfig) (*SSLDeployerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl deployer provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	// 创建 SSL 客户端 - 使用SSL专用端点
	sslClient, err := createSSLClient(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create ssl client: %w", err)
	}

	// 创建 SSL 管理器
	sslManagerConfig := &tencentcloudssl.SSLManagerProviderConfig{
		SecretId:  config.SecretId,
		SecretKey: config.SecretKey,
		Endpoint:  config.Endpoint,
	}
	sslManager, err := tencentcloudssl.NewSSLManagerProvider(sslManagerConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create ssl manager: %w", err)
	}

	return &SSLDeployerProvider{
		config:     config,
		logger:     logger.GetLogger(),
		sdkClient:  client,
		sslClient:  sslClient,
		sslManager: sslManager,
	}, nil
}

func (d *SSLDeployerProvider) Name() string {
	return "tencentcloud-cdn"
}

func (d *SSLDeployerProvider) DeployCertificate(ctx context.Context, certContent, keyContent string, workflowConfig map[string]interface{}) error {
	// 上传证书到腾讯云 SSL
	certId, err := d.uploadCertificate(ctx, certContent, keyContent)
	if err != nil {
		return fmt.Errorf("failed to upload certificate: %w", err)
	}

	request := tccdn.NewDescribeDomainsConfigRequest()
	request.Filters = []*tccdn.DomainFilter{
		&tccdn.DomainFilter{
			Name:  common.StringPtr("domain"),
			Value: common.StringPtrs([]string{d.config.Domains}),
		},
	}
	response, err := d.sdkClient.DescribeDomainsConfig(request)
	if err != nil {
		return fmt.Errorf("failed to describe domains: %w", err)
	}

	// 检查域名是否存在
	if len(response.Response.Domains) == 0 {
		return fmt.Errorf("domain %s not found", d.config.Domains)
	}

	// 获取域名配置
	domainConfig := response.Response.Domains[0]
	if domainConfig.Https.Switch == common.StringPtr("on") && domainConfig.Https.CertInfo.CertId == common.StringPtr(certId.CertId) {
		// 已配置证书，无需重复配置
		return nil
	}

	// 配置 HTTPS
	updaterequest := tccdn.NewUpdateDomainConfigRequest()
	updaterequest.Domain = common.StringPtr(d.config.Domains)
	updaterequest.Https = &tccdn.Https{
		Switch: common.StringPtr("on"),
		CertInfo: &tccdn.ServerCert{
			CertId: common.StringPtr(certId.CertId),
		},
	}
	updater, err := d.sdkClient.UpdateDomainConfig(updaterequest)
	if err != nil {
		return fmt.Errorf("failed to update domain config: %w", err)
	}
	d.logger.Debug("update domain config success", zap.Any("response", updater))

	return nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*tccdn.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tccdn.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// createSSLClient 创建 SSL 客户端
func createSSLClient(secretId, secretKey, endpoint string) (*tcssl.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	// 根据 endpoint 判断是否为国际站
	if strings.HasSuffix(endpoint, "intl.tencentcloudapi.com") {
		cpf.HttpProfile.Endpoint = "ssl.intl.tencentcloudapi.com"
	} else {
		// 强制使用SSL服务的正确端点
		cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
	}

	client, err := tcssl.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// uploadCertificate 上传证书到腾讯云 SSL
func (d *SSLDeployerProvider) uploadCertificate(ctx context.Context, certPEM string, privkeyPEM string) (*struct{ CertId string }, error) {
	// 使用 SSLManagerProvider 上传证书
	result, err := d.sslManager.Upload(ctx, certPEM, privkeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to upload certificate: %w", err)
	}

	return &struct{ CertId string }{CertId: result.CertId}, nil
}
