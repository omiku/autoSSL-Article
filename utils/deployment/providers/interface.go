package providers

import "context"

// DeployProvider 部署提供者接口定义
// 所有部署方式都需要实现此接口
type DeployProvider interface {
	// Name 返回部署提供者名称
	Name() string

	// DeployCertificate 部署证书到目标服务
	// ctx: 上下文对象
	// certContent: 证书内容字符串
	// keyContent: 私钥内容字符串
	// workflowConfig: 工作流额外配置
	DeployCertificate(ctx context.Context, certContent, keyContent string, workflowConfig map[string]interface{}) error
}

// DeployProviderRegistry 部署提供者注册器
type DeployProviderRegistry interface {
	// Register 注册部署提供者
	Register(provider DeployProvider)

	// Get 获取指定名称的部署提供者
	Get(name string) (DeployProvider, bool)

	// List 列出所有注册的部署提供者
	List() []DeployProvider
}
