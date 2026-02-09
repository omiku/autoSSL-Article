package onepanelsite

import (
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"resty.dev/v3"
)

type Client struct {
	client *resty.Client
}

// 根据文档计算访问token
func getAccessToken(apiKey string) (tokenMd5Hex, timestampStr string) {
	// 获取当前的文本类型的unix时间戳
	timestamp := time.Now().Unix()
	timestampStr = fmt.Sprintf("%d", timestamp)
	tokenMd5 := md5.Sum([]byte("1panel" + apiKey + timestampStr))
	tokenMd5Hex = hex.EncodeToString(tokenMd5[:])
	return tokenMd5Hex, timestampStr
}

// NewClient 创建1Panel HTTP客户端
func NewClient(siteURL, apiKey string) (*Client, error) {

	if siteURL == "" || apiKey == "" {
		return nil, fmt.Errorf("siteURL和apiKey不能为空")
	}

	if _, err := url.Parse(siteURL); err != nil {
		return nil, fmt.Errorf("siteURL格式错误: %v", err)
	}

	tokenMd5Hex, timestampStr := getAccessToken(apiKey)

	client := resty.New()
	client.SetBaseURL(strings.TrimRight(siteURL, "/") + "/api/v2")
	client.SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("1Panel-Token", tokenMd5Hex).
		SetHeader("1Panel-Timestamp", timestampStr)
	return &Client{client: client}, nil
}

// SetTimeout 设置请求超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.client.SetTimeout(timeout)
}

func (c *Client) SetTLSConfig(config *tls.Config) *Client {
	c.client.SetTLSClientConfig(config)
	return c
}

func (c *Client) newRequest(method string, path string) (*resty.Request, error) {
	if method == "" {
		return nil, fmt.Errorf("method不能为空")
	}
	if path == "" {
		return nil, fmt.Errorf("path不能为空")
	}

	req := c.client.R()
	req.Method = method
	req.URL = path
	return req, nil
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(req *resty.Request) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("请求对象不能为空")
	}

	resp, err := req.Send()
	if err != nil {
		// 即使有错误也返回响应对象，以便上层可以访问错误响应体
		return resp, fmt.Errorf("执行请求失败: %v", err)
	}

	return resp, nil
}

func (c *Client) doRequestWithResult(req *resty.Request, res apiResponse) (*resty.Response, error) {
	if req == nil {
		return nil, fmt.Errorf("sdkerr: nil request")
	}

	resp, err := c.doRequest(req)

	// 检查响应对象是否为空，避免空指针异常
	if resp == nil {
		return nil, err
	}

	// 安全地读取响应体
	bodyBytes, readErr := io.ReadAll(resp.Body)
	if err != nil {
		// 如果请求失败但响应体读取成功，尝试解析错误信息
		if readErr == nil && len(bodyBytes) > 0 {
			json.Unmarshal(bodyBytes, &res)
		}
		return resp, err
	}

	// 如果读取响应体失败，但请求成功
	if readErr != nil {
		return resp, fmt.Errorf("读取响应体失败: %v", readErr)
	}

	// 解析响应体
	if len(bodyBytes) != 0 {
		if err := json.Unmarshal(bodyBytes, &res); err != nil {
			return resp, fmt.Errorf("unmarshal response failed: %w", err)
		} else {
			if tcode := res.GetCode(); tcode/100 != 2 {
				return resp, fmt.Errorf("api error: code='%d', message='%s'", tcode, res.GetMessage())
			}
		}
	}

	return resp, nil
}

func (c *Client) GetHttpsConf(websiteId int64) (*GetHttpsConfResponse, error) {
	return c.GetHttpsConfWithContext(context.Background(), websiteId)
}

func (c *Client) GetHttpsConfWithContext(ctx context.Context, websiteId int64) (*GetHttpsConfResponse, error) {
	if websiteId == 0 {
		return nil, fmt.Errorf("unset websiteId")
	}

	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/websites/%d/https", websiteId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &GetHttpsConfResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}

func (c *Client) UpdateHttpsConf(websiteId int64, req *UpdateHttpsConfRequest) (*UpdateHttpsConfResponse, error) {
	return c.UpdateHttpsConfWithContext(context.Background(), websiteId, req)
}

func (c *Client) UpdateHttpsConfWithContext(ctx context.Context, websiteId int64, req *UpdateHttpsConfRequest) (*UpdateHttpsConfResponse, error) {
	if websiteId == 0 {
		return nil, fmt.Errorf("sdkerr: unset websiteId")
	}

	httpreq, err := c.newRequest(http.MethodPost, fmt.Sprintf("/websites/%d/https", websiteId))
	if err != nil {
		return nil, err
	} else {
		req.WebsiteID = websiteId
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpdateHttpsConfResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}

// UploadWebsiteSSL 上传网站SSL证书
func (c *Client) UploadWebsiteSSL(req *UploadWebsiteSSLRequest) (*UploadWebsiteSSLResponse, error) {
	return c.UploadWebsiteSSLWithContext(context.Background(), req)
}

func (c *Client) UploadWebsiteSSLWithContext(ctx context.Context, req *UploadWebsiteSSLRequest) (*UploadWebsiteSSLResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/websites/ssl/upload")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UploadWebsiteSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}

func (c *Client) SearchWebsite(req *SearchWebsiteRequest) (*SearchWebsiteResponse, error) {
	return c.SearchWebsiteWithContext(context.Background(), req)
}

func (c *Client) SearchWebsiteWithContext(ctx context.Context, req *SearchWebsiteRequest) (*SearchWebsiteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("unset request")
	}

	httpreq, err := c.newRequest(http.MethodPost, "/websites/search")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &SearchWebsiteResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}

// SearchWebsiteSSL 搜索网站SSL证书
func (c *Client) SearchWebsiteSSL(req *SearchWebsiteSSLRequest) (*SearchWebsiteSSLResponse, error) {
	return c.SearchWebsiteSSLWithContext(context.Background(), req)
}

func (c *Client) SearchWebsiteSSLWithContext(ctx context.Context, req *SearchWebsiteSSLRequest) (*SearchWebsiteSSLResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/websites/ssl/search")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &SearchWebsiteSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}

// UpdateCoreSettingsSSL 更新核心设置SSL证书
func (c *Client) UpdateCoreSettingsSSL(req *UpdateCoreSettingsSSLRequest) (*UpdateCoreSettingsSSLResponse, error) {
	return c.UpdateCoreSettingsSSLWithContext(context.Background(), req)
}

func (c *Client) UpdateCoreSettingsSSLWithContext(ctx context.Context, req *UpdateCoreSettingsSSLRequest) (*UpdateCoreSettingsSSLResponse, error) {
	httpreq, err := c.newRequest(http.MethodPost, "/core/settings/ssl")
	if err != nil {
		return nil, err
	} else {
		httpreq.SetBody(req)
		httpreq.SetContext(ctx)
	}

	result := &UpdateCoreSettingsSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}

func (c *Client) GetWebsiteSSL(sslId int64) (*GetWebsiteSSLResponse, error) {
	return c.GetWebsiteSSLWithContext(context.Background(), sslId)
}

func (c *Client) GetWebsiteSSLWithContext(ctx context.Context, sslId int64) (*GetWebsiteSSLResponse, error) {
	if sslId == 0 {
		return nil, fmt.Errorf("unset sslId")
	}

	httpreq, err := c.newRequest(http.MethodGet, fmt.Sprintf("/websites/ssl/%d", sslId))
	if err != nil {
		return nil, err
	} else {
		httpreq.SetContext(ctx)
	}

	result := &GetWebsiteSSLResponse{}
	if _, err := c.doRequestWithResult(httpreq, result); err != nil {
		return result, err
	}

	return result, nil
}
