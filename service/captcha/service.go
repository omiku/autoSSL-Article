package captcha

import (
	"autoSSL/config"
	"sync"
	"time"
)

type Service struct {
	adapter CaptchaAdapter
	cache   sync.Map
	config  config.CaptchaConfig
}

func New(config config.CaptchaConfig) (*Service, error) {
	adapter, err := NewCaptchaFactory(config)
	if err != nil {
		return nil, err
	}

	return &Service{
		adapter: adapter,
		config:  config,
	}, nil
}

func (s *Service) Generate() (string, Base64Adapter, error) {
	id, data, err := s.adapter.Generate()
	if err != nil {
		return "", Base64Adapter{}, err
	}

	// 自动清理缓存（根据配置设置超时）
	expire := time.Minute * time.Duration(s.config.ExpireMinutes)
	time.AfterFunc(expire, func() { s.cache.Delete(id) })

	return id, data, nil
}

func (s *Service) Verify(id, answer string) bool {
	return s.adapter.Verify(id, answer)
}
