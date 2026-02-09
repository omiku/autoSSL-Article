package providers

import (
	"sync"
)

// registry 部署提供者注册器的默认实现
type registry struct {
	providers map[string]DeployProvider
	mu        sync.RWMutex
}

// NewDeployProviderRegistry 创建新的部署提供者注册器
func NewDeployProviderRegistry() DeployProviderRegistry {
	return &registry{
		providers: make(map[string]DeployProvider),
	}
}

// Register 注册部署提供者
func (r *registry) Register(provider DeployProvider) {
	if provider == nil {
		return
	}

	name := provider.Name()
	if name == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[name] = provider
}

// Get 获取指定名称的部署提供者
func (r *registry) Get(name string) (DeployProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[name]
	return provider, ok
}

// List 列出所有注册的部署提供者
func (r *registry) List() []DeployProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]DeployProvider, 0, len(r.providers))
	for _, provider := range r.providers {
		list = append(list, provider)
	}

	return list
}

// 全局默认注册器实例
var defaultRegistry = NewDeployProviderRegistry()

// DefaultRegistry 获取默认注册器
func DefaultRegistry() DeployProviderRegistry {
	return defaultRegistry
}

// RegisterProvider 注册部署提供者到默认注册器
func RegisterProvider(provider DeployProvider) {
	defaultRegistry.Register(provider)
}
