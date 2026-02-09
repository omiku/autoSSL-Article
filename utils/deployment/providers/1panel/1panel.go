package onepanelsite

import (
	"autoSSL/logger"
	"autoSSL/utils/deployment/providers"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

//MKcI1GF1aB0M1ZqvOD2ClQU7HyvYuq1j

// OnePanelSite 1Panel站点配置
type OnePanelDeployerConfig struct {
	SiteURL                  string `json:"siteUrl"`
	WebsiteID                int64  `json:"websiteId"`
	CertificateID            int64  `json:"certificateId"`
	ResourceType             string `json:"resourceType"`
	AllowInsecureConnections bool   `json:"allowInsecureConnections,omitempty"`
	ApiKey                   string `json:"apiKey"`
	Domain                   string `json:"domain"`
}

// OnePanelDeployerProvider 1Panel部署器提供程序
type OnePanelDeployer struct {
	config *OnePanelDeployerConfig
	client *Client
	logger *logger.Logger
}

var _ providers.DeployProvider = (*OnePanelDeployer)(nil)

// NewOnePanelDeployer 创建1Panel部署器
func NewOnePanelDeployer(config *OnePanelDeployerConfig) (*OnePanelDeployer, error) {
	if config == nil {
		return nil, errors.New("the configuration of the 1panel deployer provider is nil")
	}
	client, err := NewClient(config.SiteURL, config.ApiKey)
	if err != nil {
		return nil, err
	}

	if config.AllowInsecureConnections {
		client.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return &OnePanelDeployer{
		config: config,
		logger: logger.GetLogger(),
		client: client,
	}, nil
}

func (d *OnePanelDeployer) Name() string {
	return "1panel"
}

func (d *OnePanelDeployer) DeployCertificate(ctx context.Context, certContent, keyContent string, workflowConfig map[string]interface{}) error {
	switch d.config.ResourceType {
	case "website":
		return d.DeployToWebsite(ctx, certContent, keyContent)
	case "certificate":
		return d.DeployToCertificate(ctx, certContent, keyContent)
	default:
		return errors.New("invalid resource type")
	}
	// return nil
}

// 部署给指定网站ID的证书
func (d *OnePanelDeployer) DeployToWebsite(ctx context.Context, certContent, keyContent string) error {
	if d.config.WebsiteID == 0 || d.config.Domain == "" {
		return errors.New("unset websiteId or domain")
	}

	// upload certificate
	certID, err := d.uploadCertificate(ctx, certContent, keyContent)
	if err != nil {
		return fmt.Errorf("upload certificate failed: %w", err)
	}

	// 获取网站ID
	if d.config.WebsiteID == 0 {
		websiteID, err := d.getWebsiteID(ctx, d.config.Domain)
		if err != nil {
			return fmt.Errorf("get website id failed: %w", err)
		}
		d.config.WebsiteID = websiteID
	}

	// 拿网站配置
	gethttpsConf, err := d.client.GetHttpsConf(d.config.WebsiteID)
	if err != nil {
		return fmt.Errorf("get website ssl failed: %w", err)
	}

	// 修改网站https配置
	updateHttpsConfReq := &UpdateHttpsConfRequest{
		Enable:       gethttpsConf.Data.Enable,
		Type:         "existed",
		WebsiteID:    d.config.WebsiteID,
		WebsiteSSLID: certID,
		HttpConfig:   gethttpsConf.Data.HttpConfig,
		SSLProtocol:  gethttpsConf.Data.SSLProtocol,
		Algorithm:    gethttpsConf.Data.Algorithm,
		Hsts:         gethttpsConf.Data.Hsts,
	}
	_, err = d.client.UpdateHttpsConf(d.config.WebsiteID, updateHttpsConfReq)
	if err != nil {
		return fmt.Errorf("update website ssl failed: %w", err)
	}

	return nil
}

// 部署指定的证书id
func (d *OnePanelDeployer) DeployToCertificate(ctx context.Context, certContent, keyContent string) error {
	if d.config.CertificateID == 0 {
		return errors.New("unset certificateId")
	}

	// 获取证书详情
	cert, err := d.client.GetWebsiteSSL(d.config.CertificateID)
	if err != nil {
		return fmt.Errorf("get certificate failed: %w", err)
	}

	// 拼装更新证书请求
	updateCertReq := &UploadWebsiteSSLRequest{
		SSLID:       d.config.CertificateID,
		Type:        "paste",
		Description: cert.Data.Description,
		Certificate: certContent,
		PrivateKey:  keyContent,
	}

	// 更新证书
	uploadWebsiteSSLResp, err := d.client.UploadWebsiteSSL(updateCertReq)
	d.logger.Debug("update certificate response", zap.Any("response", uploadWebsiteSSLResp))
	if err != nil {
		return fmt.Errorf("update certificate failed: %w", err)
	}

	return nil
}

func (d *OnePanelDeployer) GetHttpsConf(ctx context.Context) (*GetHttpsConfResponse, error) {
	return d.client.GetHttpsConfWithContext(ctx, d.config.WebsiteID)
}

// 上传证书
func (d *OnePanelDeployer) uploadCertificate(ctx context.Context, certContent, keyContent string) (int64, error) {
	// check certificate exists
	certID, err := d.checkCertificateExists(ctx, certContent, keyContent)
	if err != nil {
		return 0, fmt.Errorf("check certificate exists failed: %w", err)
	}
	if certID > 0 {
		d.logger.Debug("certificate %d already exists", zap.Int64("certID", certID))
		return certID, fmt.Errorf("certificate %d already exists", certID)
	}
	req := &UploadWebsiteSSLRequest{
		Type:        "paste",
		Description: d.config.Domain,
		Certificate: certContent,
		PrivateKey:  keyContent,
	}
	_, err = d.client.UploadWebsiteSSL(req)
	if err != nil {
		return 0, fmt.Errorf("upload certificate failed: %w", err)
	}
	// 上传后获取id
	certID, err = d.checkCertificateExists(ctx, certContent, keyContent)
	if err != nil {
		return 0, fmt.Errorf("check certificate exists failed: %w", err)
	}
	return certID, nil
}

// 根据证书内容检查证书是否已存在
func (d *OnePanelDeployer) checkCertificateExists(ctx context.Context, certContent, keyContent string) (int64, error) {
	// 从1Panel获取所有证书
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
		}
		searchWebsiteSSLPageNumber := int32(1)
		searchWebsiteSSLPageSize := int32(100)
		searchWebsiteSSLItemsCount := int32(0)
		searchWebsiteSSLReq := &SearchWebsiteSSLRequest{
			Page:     searchWebsiteSSLPageNumber,
			PageSize: searchWebsiteSSLPageSize,
		}
		certs, err := d.client.SearchWebsiteSSL(searchWebsiteSSLReq)
		if err != nil {
			return 0, fmt.Errorf("list certificates failed: %w", err)
		}

		// 遍历所有证书，检查是否有匹配的
		if certs.Data != nil {
			for _, sslItem := range certs.Data.Items {
				if strings.TrimSpace(sslItem.PEM) == strings.TrimSpace(certContent) &&
					strings.TrimSpace(sslItem.PrivateKey) == strings.TrimSpace(keyContent) {
					// 如果已存在相同证书，直接返回
					return sslItem.ID, nil
				}
			}
		}

		searchWebsiteSSLItemsCount = certs.Data.Total
		if searchWebsiteSSLItemsCount < searchWebsiteSSLPageSize {
			break
		} else {
			searchWebsiteSSLPageNumber++
		}
	}

	return 0, nil
}

// 根据单行domain获取websiteId
func (d *OnePanelDeployer) getWebsiteID(ctx context.Context, domain string) (int64, error) {

	searchWebsiteReq := &SearchWebsiteRequest{
		Name:           domain,
		Order:          "descending",
		Page:           1,
		PageSize:       20,
		Type:           "favorite",
		WebsiteGroupID: 0,
	}
	websites, err := d.client.SearchWebsite(searchWebsiteReq)
	if err != nil {
		return 0, fmt.Errorf("list websites failed: %w", err)
	}
	if websites.Data != nil && len(websites.Data.Items) > 0 {
		d.logger.Debug("website %s found, id: %d", zap.String("domain", domain), zap.Int64("id", websites.Data.Items[0].ID))
		return websites.Data.Items[0].ID, nil
	}

	return 0, fmt.Errorf("website %s not found", domain)
}
