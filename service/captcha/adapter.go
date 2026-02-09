package captcha

import (
	"autoSSL/config"
)

type CaptchaAdapter interface {
	Generate() (id string, imageData Base64Adapter, err error)
	Verify(id string, answer string) bool
}

type Config struct {
	Type      string
	Base64    config.Base64Config
	GoCaptcha config.GoCaptchaConfig
}

// 为两个验证库创建兼容存储bs64结构体
type Base64Adapter struct {
	NormBase64, MasterBase64, ThumbBase64 string
}

// type Base64Config struct {
// 	DigitHeight int
// 	DigitWidth  int
// 	DigitLength int
// 	MaxSkew     float64
// 	DotCount    int
// }

// type GoCaptchaConfig struct {
// 	ClickLen        int
// 	ClickRangeSize  int
// 	BackgroundImage string
// }
