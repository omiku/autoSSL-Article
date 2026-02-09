package captcha

import (
	"autoSSL/config"
	"sync"
	"time"

	"github.com/mojocn/base64Captcha"
)

type Base64CaptchaAdapter struct {
	config config.Base64Config
	store  base64Captcha.Store
	cache  sync.Map
}

func NewBase64Captcha(config config.Base64Config) *Base64CaptchaAdapter {
	return &Base64CaptchaAdapter{
		config: config,
		store:  base64Captcha.DefaultMemStore,
	}
}

func (a *Base64CaptchaAdapter) Generate() (id string, b64ad Base64Adapter, err error) {
	driver := base64Captcha.NewDriverDigit(
		a.config.DigitHeight,
		a.config.DigitWidth,
		a.config.DigitLength,
		a.config.MaxSkew,
		a.config.DotCount,
	)
	cp := base64Captcha.NewCaptcha(driver, a.store)
	id, b64s, answer, err := cp.Generate()
	if err != nil {
		return "", b64ad, err
	}
	a.cache.Store(id, answer)
	time.AfterFunc(5*time.Minute, func() { a.cache.Delete(id) })
	b64ad.NormBase64 = b64s
	return id, b64ad, nil
}

func (a *Base64CaptchaAdapter) Verify(id, answer string) bool {
	v, ok := a.cache.Load(id)
	if !ok {
		return false
	}
	a.cache.Delete(id)
	return v.(string) == answer
}
