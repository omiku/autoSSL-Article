package kuocai_cdn

import (
	"autoSSL/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LoginResponse 登录返回的结构体
type LoginResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// ListDomainsResponse 列出加速域名返回的结构体
type ListDomainsResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		RecordsFiltered int `json:"recordsFiltered"`
		Data            []struct {
			Id         string `json:"id"`
			DomainName string `json:"domainName"`
		} `json:"data"`
		RecordsTotal int `json:"recordsTotal"`
	} `json:"data"`
	Success             bool `json:"success"`
	SuccessWithDateResp bool `json:"successWithDateResp"`
}

// KuocaiClient 可复用的HTTP客户端
type KuocaiClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
	logger     *logger.Logger
}

// NewKuocaiClient 创建新的客户端实例
func NewKuocaiClient() *KuocaiClient {
	return &KuocaiClient{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		baseURL: "https://kuocai.cn",
		logger:  logger.GetLogger(),
	}
}

// Login 登录并保存token
func (c *KuocaiClient) Login(account, password string) error {
	loginData := fmt.Sprintf("userAccount=%s&userPwd=%s&remember=false",
		account, password)

	resp, err := c.Post("/login/loginUser", loginData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return err
	}

	if loginResp.Code != "SUCCESS" {
		return fmt.Errorf("login failed: %s", loginResp.Message)
	}

	c.token = loginResp.Data
	return nil
}

// Post 发送POST请求（自动添加认证头）
func (c *KuocaiClient) Post(path string, body string) (*http.Response, error) {
	url := c.baseURL + path
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// 设置通用头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")

	// 如果已登录，添加认证cookie
	if c.token != "" {
		req.Header.Set("Cookie", fmt.Sprintf("kuocai_cdn_token=%s", c.token))
	}

	return c.httpClient.Do(req)
}

// Get 发送GET请求（自动添加认证头）
func (c *KuocaiClient) Get(path string) (*http.Response, error) {
	url := c.baseURL + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置通用头
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")

	// 如果已登录，添加认证cookie
	if c.token != "" {
		req.Header.Set("Cookie", fmt.Sprintf("kuocai_cdn_token=%s", c.token))
	}

	return c.httpClient.Do(req)
}
