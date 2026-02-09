package tencentcloudteo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
	tcteo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"

	"go.uber.org/zap"

	"autoSSL/logger"
	"autoSSL/utils/deployment/providers"
	"autoSSL/utils/deployment/providers/tencentcloudssl"
)

type SSLDeployerProviderConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
	// 站点 ID。
	ZoneId string `json:"zoneId"`
	// 加速域名列表（支持泛域名）。
	Domains []string `json:"domains"`
}

type SSLDeployerProvider struct {
	config     *SSLDeployerProviderConfig
	logger     *logger.Logger
	sdkClient  *tcteo.Client
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
	return "tencentcloud-teo"
}

func (d *SSLDeployerProvider) DeployCertificate(ctx context.Context, certContent, keyContent string, workflowConfig map[string]interface{}) error {
	if d.config.ZoneId == "" {
		return errors.New("config `zoneId` is required")
	}

	// 从workflowConfig中获取域名
	var domains []string
	if len(d.config.Domains) > 0 {
		domains = d.config.Domains
	} else if domain, ok := workflowConfig["domain"].(string); ok && domain != "" {
		domains = []string{domain}
	} else {
		return errors.New("config `domains` is required")
	}

	// 上传证书
	upres, err := d.uploadCertificate(ctx, certContent, keyContent)
	if err != nil {
		return fmt.Errorf("failed to upload certificate file: %w", err)
	} else {
		d.logger.Info("ssl certificate uploaded", zap.Any("result", upres))
	}

	// 配置域名证书
	// REF: https://cloud.tencent.com/document/api/1552/80764
	modifyHostsCertificateReq := tcteo.NewModifyHostsCertificateRequest()
	modifyHostsCertificateReq.ZoneId = common.StringPtr(d.config.ZoneId)
	modifyHostsCertificateReq.Mode = common.StringPtr("sslcert")
	modifyHostsCertificateReq.Hosts = common.StringPtrs(domains)
	modifyHostsCertificateReq.ServerCertInfo = []*tcteo.ServerCertInfo{{CertId: common.StringPtr(upres.CertId)}}
	modifyHostsCertificateResp, err := d.sdkClient.ModifyHostsCertificate(modifyHostsCertificateReq)
	d.logger.Debug("sdk request 'teo.ModifyHostsCertificate'", zap.Any("request", modifyHostsCertificateReq), zap.Any("response", modifyHostsCertificateResp))
	if err != nil {
		return fmt.Errorf("failed to execute sdk request 'teo.ModifyHostsCertificate': %w", err)
	}

	return nil
}

func createSDKClient(secretId, secretKey, endpoint string) (*tcteo.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	if endpoint != "" {
		cpf.HttpProfile.Endpoint = endpoint
	}

	client, err := tcteo.NewClient(credential, "", cpf)
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
