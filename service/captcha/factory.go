package captcha

import (
	"autoSSL/config"
	"errors"
)

func NewCaptchaFactory(conf config.CaptchaConfig) (CaptchaAdapter, error) {
	switch conf.Type {
	case "base64captcha":
		return NewBase64Captcha(conf.Base64), nil
	case "gocaptcha":
		return NewGoCaptcha(conf.GoCaptcha)
	default:
		return nil, errors.New("unsupported captcha type")
	}
}
