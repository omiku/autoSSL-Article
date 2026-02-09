package deployer

import (
	"autoSSL/ent"
	"autoSSL/logger"
	"autoSSL/utils/deployment/providers"
	"context"
	"fmt"
)

// Deployer 部署器接口
type Deployer interface {
	Deploy(ctx context.Context) error
}

// deployerImpl 部署器实现
type deployerImpl struct {
	provider              providers.DeployProvider
	certPEM               string
	privkeyPEM            string
	ProviderServiceConfig map[string]interface{}
}

// DeploymentService 部署服务协调器
type DeploymentService struct {
	registry providers.DeployProviderRegistry
}

// DeployerWithWorkflowAccessConfig 工作流节点部署配置
type DeployerWithWorkflowAccessConfig struct {
	Access                *ent.Access
	Logger                *logger.Logger
	CertificatePEM        string
	PrivateKeyPEM         string
	ProviderServiceConfig map[string]interface{}
}

// NewWithWorkflowAccess 根据工作流节点配置创建部署器
func NewWithWorkflowAccess(config DeployerWithWorkflowAccessConfig) (Deployer, error) {
	if config.Access == nil {
		return nil, fmt.Errorf("access is nil")
	}

	providerName := config.Access.Type
	if providerName == "" {
		return nil, fmt.Errorf("provider name is empty")
	}

	// 准备部署提供者选项
	options := &deployerProviderOptions{
		Provider:              providerName.String(),
		ProviderAccessConfig:  config.Access.Config,
		ProviderServiceConfig: config.ProviderServiceConfig,
	}

	// 创建部署提供者
	provider, err := createDeployerProvider(options)
	if err != nil {
		return nil, fmt.Errorf("创建部署提供者失败: %v, 配置: %v", err, options)
	}

	return &deployerImpl{
		provider:              provider,
		certPEM:               config.CertificatePEM,
		privkeyPEM:            config.PrivateKeyPEM,
		ProviderServiceConfig: config.ProviderServiceConfig,
	}, nil
}

// Deploy 执行部署操作
func (d *deployerImpl) Deploy(ctx context.Context) error {
	logger.Info(fmt.Sprintf("使用部署提供者 '%s' 部署证书", d.provider.Name()))

	// 执行部署
	if err := d.provider.DeployCertificate(ctx, d.certPEM, d.privkeyPEM, d.ProviderServiceConfig); err != nil {
		logger.Error(fmt.Sprintf("部署证书失败: %v", err))
		return fmt.Errorf("部署证书失败: %w", err)
	}

	logger.Info("证书部署成功")
	return nil
}

// NewDeploymentService 创建新的部署服务实例
//func NewDeploymentService(registry providers.DeployProviderRegistry) *DeploymentService {
//	if registry == nil {
//		registry = providers.DefaultRegistry()
//	}
//
//	return &DeploymentService{
//		registry: registry,
//	}
//}

// ListProviders 列出所有可用的部署提供者
func (d *DeploymentService) ListProviders() []string {
	providerList := d.registry.List()
	names := make([]string, len(providerList))

	for i, p := range providerList {
		names[i] = p.Name()
	}

	return names
}
