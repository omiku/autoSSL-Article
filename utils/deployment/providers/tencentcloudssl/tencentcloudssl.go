package tencentcloudssl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tcssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"

	"autoSSL/logger"

	"go.uber.org/zap"
)

type SSLManagerProviderConfig struct {
	// 腾讯云 SecretId。
	SecretId string `json:"secretId"`
	// 腾讯云 SecretKey。
	SecretKey string `json:"secretKey"`
	// 腾讯云接口端点。
	Endpoint string `json:"endpoint,omitempty"`
}

type SSLManagerProvider struct {
	config    *SSLManagerProviderConfig
	logger    *logger.Logger
	sdkClient *tcssl.Client
}

func NewSSLManagerProvider(config *SSLManagerProviderConfig) (*SSLManagerProvider, error) {
	if config == nil {
		return nil, errors.New("the configuration of the ssl manager provider is nil")
	}

	client, err := createSDKClient(config.SecretId, config.SecretKey, config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("could not create sdk client: %w", err)
	}

	return &SSLManagerProvider{
		config:    config,
		logger:    logger.GetLogger(),
		sdkClient: client,
	}, nil
}

// SSLManageUploadResult 上传证书结果
type SSLManageUploadResult struct {
	CertId string
}

// SSLManagerProviderInterface 定义了 SSLManagerProvider 的接口
type SSLManagerProviderInterface interface {
	Upload(ctx context.Context, certPEM string, privkeyPEM string) (*SSLManageUploadResult, error)
}

// 确保 SSLManagerProvider 实现了 SSLManagerProviderInterface
var _ SSLManagerProviderInterface = (*SSLManagerProvider)(nil)

func (m *SSLManagerProvider) Upload(ctx context.Context, certPEM string, privkeyPEM string) (*SSLManageUploadResult, error) {
	// 上传新证书
	// REF: https://cloud.tencent.com/document/api/400/41665
	
	// 记录调试信息
	if m.logger != nil {
		m.logger.Info("开始上传证书到腾讯云SSL服务", 
			zap.Int("certLength", len(certPEM)), 
			zap.Int("keyLength", len(privkeyPEM)),
			zap.String("endpoint", m.getEndpoint()))
	}
	
	uploadCertificateReq := tcssl.NewUploadCertificateRequest()
	uploadCertificateReq.CertificatePublicKey = common.StringPtr(certPEM)
	uploadCertificateReq.CertificatePrivateKey = common.StringPtr(privkeyPEM)
	uploadCertificateReq.Repeatable = common.BoolPtr(false)
	
	uploadCertificateResp, err := m.sdkClient.UploadCertificate(uploadCertificateReq)

	// 检查 logger 是否为 nil，避免空指针引用
	if m.logger != nil {
		m.logger.Debug("SSL上传请求详情", 
			zap.Any("request", uploadCertificateReq), 
			zap.Any("response", uploadCertificateResp),
			zap.Error(err))
	}

	if err != nil {
		if m.logger != nil {
			m.logger.Error("SSL证书上传失败", 
				zap.Error(err),
				zap.String("endpoint", m.getEndpoint()))
		}
		return nil, fmt.Errorf("failed to execute sdk request 'ssl.UploadCertificate': %w", err)
	}

	if uploadCertificateResp == nil || uploadCertificateResp.Response == nil || uploadCertificateResp.Response.CertificateId == nil {
		return nil, errors.New("invalid response from SSL upload API")
	}

	certId := *uploadCertificateResp.Response.CertificateId
	if m.logger != nil {
		m.logger.Info("SSL证书上传成功", zap.String("certId", certId))
	}

	return &SSLManageUploadResult{
		CertId: certId,
	}, nil
}

// getEndpoint 获取当前使用的端点
func (m *SSLManagerProvider) getEndpoint() string {
	// 这里我们假设客户端使用的是默认端点
	return "ssl.tencentcloudapi.com"
}

func createSDKClient(secretId, secretKey, endpoint string) (*tcssl.Client, error) {
	credential := common.NewCredential(secretId, secretKey)

	cpf := profile.NewClientProfile()
	// 强制使用SSL服务的正确端点，避免错误的服务调用
	if endpoint != "" {
		// 检查是否为国际站端点
		if strings.Contains(endpoint, "intl") {
			cpf.HttpProfile.Endpoint = "ssl.intl.tencentcloudapi.com"
		} else {
			cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
		}
	} else {
		// 默认使用中国大陆SSL服务端点
		cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
	}

	client, err := tcssl.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return client, nil
}
