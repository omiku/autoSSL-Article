package captcha

import (
	"autoSSL/config"
	"crypto/rand"
	"encoding/json"
	"image"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/wenlng/go-captcha-assets/bindata/chars"
	assets "github.com/wenlng/go-captcha-assets/bindata/images/image_v2_1"
	"github.com/wenlng/go-captcha-assets/resources/fonts/fzshengsksjw"
	"github.com/wenlng/go-captcha/v2/base/codec"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/click"
)

type GoCaptchaAdapter struct {
	config  config.GoCaptchaConfig
	builder click.Builder
	cache   sync.Map
}

func NewGoCaptcha(config config.GoCaptchaConfig) (*GoCaptchaAdapter, error) {
	builder := click.NewBuilder(
		click.WithRangeLen(option.RangeVal{Min: config.ClickLen, Max: config.ClickLen + 2}),
		click.WithRangeVerifyLen(option.RangeVal{Min: 2, Max: config.ClickRangeSize}),
	)
	fontBytes, err := fzshengsksjw.GetFont()
	if err != nil {
		log.Printf("加载预置字体失败: %v", err)
		return nil, err
	}

	bgAsset, err := assets.Asset(config.BackgroundImage)
	if err != nil {
		log.Printf("加载背景图资源失败: %v", err)
		return nil, err
	}
	bgImage, err := codec.DecodeByteToJpeg(bgAsset)
	if err != nil {
		log.Printf("解码背景图失败: %v", err)
		return nil, err
	}
	builder.SetResources(
		click.WithChars(chars.GetChineseChars()),
		click.WithFonts([]*truetype.Font{fontBytes}),
		click.WithBackgrounds([]image.Image{bgImage}),
	)

	return &GoCaptchaAdapter{
		config:  config,
		builder: builder,
	}, nil
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// 如果加密随机数生成失败，回退到时间种子随机数
			b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		} else {
			b[i] = charset[n.Int64()]
		}
		// b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 修复缺失参数类型的问题，为 b64ad 补上类型
func (a *GoCaptchaAdapter) Generate() (id string, b64ad Base64Adapter, err error) {
	captData := a.builder.Make()

	data, err := captData.Generate()
	if err != nil {
		return "", b64ad, err
	}

	id = generateRandomString(12)
	answer := getClickPositions(data.GetData())
	a.cache.Store(id, answer)
	time.AfterFunc(5*time.Minute, func() { a.cache.Delete(id) })

	b64ad.MasterBase64, err = data.GetMasterImage().ToBase64()
	if err != nil {
		return "", b64ad, err
	}

	b64ad.ThumbBase64, err = data.GetThumbImage().ToBase64()
	if err != nil {
		return "", b64ad, err
	}

	return id, b64ad, nil
}

func (a *GoCaptchaAdapter) Verify(id, answer string) bool {
	v, ok := a.cache.Load(id)
	if !ok {
		return false
	}
	a.cache.Delete(id)
	return v.(string) == answer
}

func getClickPositions(data map[int]*click.Dot) string {
	positions := make([]struct {
		X int `json:"x"`
		Y int `json:"y"`
	}, 0)
	for _, dot := range data {
		positions = append(positions, struct {
			X int `json:"x"`
			Y int `json:"y"`
		}{
			X: dot.X,
			Y: dot.Y,
		})
	}
	jsonData, _ := json.Marshal(positions)
	return string(jsonData)
}
