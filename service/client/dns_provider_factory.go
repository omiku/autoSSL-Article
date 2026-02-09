package client

import (
	"fmt"

	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/providers/dns/alidns"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/providers/dns/tencentcloud"
)

type DNSProviderFactory struct {
	ak      string
	sk      string
	apiToken string
}

func NewDNSProviderFactory(ak, sk, apiToken string) *DNSProviderFactory {
	return &DNSProviderFactory{
		ak:       ak,
		sk:       sk,
		apiToken: apiToken,
	}
}

func (f *DNSProviderFactory) CreateProvider(providerName string) (challenge.Provider, error) {
	switch providerName {
	case "tencentcloud":
		cfg := tencentcloud.NewDefaultConfig()
		cfg.SecretID = f.ak
		cfg.SecretKey = f.sk
		return tencentcloud.NewDNSProviderConfig(cfg)
	case "cloudflare":
		cfg := cloudflare.NewDefaultConfig()
		// 优先使用API Token，如果没有则使用API Key
		if f.apiToken != "" {
			cfg.AuthToken = f.apiToken
		} else {
			cfg.AuthEmail = f.ak
			cfg.AuthKey = f.sk
		}
		return cloudflare.NewDNSProviderConfig(cfg)
	case "aliyun":
		cfg := alidns.NewDefaultConfig()
		cfg.APIKey = f.ak
		cfg.SecretKey = f.sk
		return alidns.NewDNSProviderConfig(cfg)
	default:
		return nil, fmt.Errorf("unsupported DNS provider: %s", providerName)
	}
}
