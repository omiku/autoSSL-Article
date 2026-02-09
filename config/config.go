package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
	Captcha  CaptchaConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Driver string
	// 详细连接参数
	Host       string
	Port       int
	User       string
	Password   string
	Name       string
	UnixSocket bool `mapstructure:"unix_socket"`
}

type LoggerConfig struct {
	Level       string `mapstructure:"level"`      // 日志级别: debug/info/warn/error
	Filename    string `mapstructure:"filename"`   // 日志文件路径，为空时输出到控制台
	MaxSize     int    `mapstructure:"max_size"`   // 日志文件最大大小(MB)
	MaxBackups  int    `mapstructure:"max_backups"` // 保留的旧日志文件最大数量
	MaxAge      int    `mapstructure:"max_age"`      // 保留旧日志文件的最大天数
	Compress    bool   `mapstructure:"compress"`   // 是否压缩/归档旧日志文件
	Development bool   `mapstructure:"development"` // 开发模式，影响日志格式
}

// 保留现有验证码配置
type CaptchaConfig struct {
	Type          string          `mapstructure:"type"`
	Base64        Base64Config    `mapstructure:"base64"`
	GoCaptcha     GoCaptchaConfig `mapstructure:"gocaptcha"`
	ExpireMinutes time.Duration
}

type Base64Config struct {
	DigitHeight int     `mapstructure:"digit_height"`
	DigitWidth  int     `mapstructure:"digit_width"`
	DigitLength int     `mapstructure:"digit_length"`
	MaxSkew     float64 `mapstructure:"max_skew"`
	DotCount    int     `mapstructure:"dot_count"`
}

type GoCaptchaConfig struct {
	Store           string
	ClickLen        int
	ClickRangeSize  int
	BackgroundImage string
}

type JWTConfig struct {
	Secret        string
	Issuer        string
	AccessExpire  time.Duration `mapstructure:"access_expire"`
	RefreshExpire time.Duration `mapstructure:"refresh_expire"`
}

func LoadConfig(path string) (*AppConfig, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	log.Printf("解析后的配置: %+v", cfg)
	return &cfg, nil
}
