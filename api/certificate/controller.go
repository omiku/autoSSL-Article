package certificate

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"autoSSL/api/response"
	"autoSSL/ent"
	"autoSSL/logger"
	"autoSSL/service/certificate"
	"autoSSL/service/notification"
	"autoSSL/service/workflow"
	"autoSSL/utils/certificateutils"
	"autoSSL/utils/pagination"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CertificateController 证书管理控制器
//
//	@Tags	证书管理
type CertificateController struct {
	client *certificate.CertificateService
}

func NewCertificateController(client *certificate.CertificateService) *CertificateController {
	return &CertificateController{client: client}
}

// CreateCertificateRequest 创建证书请求参数
// @Description 创建SSL证书所需的完整请求参数
// @Example {"domains":["example.com","www.example.com","api.example.com"],"issued_by":"letsencrypt","dns_provider_id":1,"validation_type":"dns","acme_account_email":"admin@example.com","renewal_days_before":30,"access_ids":[1],"extra_config":{"remote_cert_path":"/etc/nginx/ssl/cert.pem","remote_key_path":"/etc/nginx/ssl/key.pem","pre_command":"nginx -t","post_command":"systemctl reload nginx"},"notification_emails":["admin@example.com"],"notify_on_success":true,"renewal_enabled":true}
type CreateCertificateRequest struct {
	// 域名列表
	// @Description 需要申请证书的域名列表，必须是有效的域名格式，支持多个域名合并到同一证书
	// @Example ["example.com","www.example.com","api.example.com"]
	Domains []string `json:"domains" binding:"required,min=1"`

	// 颁发机构
	// @Description 证书颁发机构，支持 letsencrypt, zerossl 等
	// @Example letsencrypt
	IssuedBy string `json:"issued_by" binding:"required"`

	// DNS提供商ID
	// @Description DNS提供商的ID，用于DNS验证
	// @Example 1
	DNSProviderID int `json:"dns_provider_id" binding:"required"`

	// 验证类型
	// @Description 域名验证方式，支持dns或http验证
	// @Enum dns,http
	// @Example dns
	ValidationType string `json:"validation_type" binding:"required,oneof=dns http"`

	// 续期提前天数
	// @Description 证书到期前多少天开始自动续期，范围1-30天
	// @Minimum 1
	// @Maximum 30
	// @Example 30
	RenewalDaysBefore int `json:"renewal_days_before" binding:"min=1,max=30"`

	// 访问配置ID列表
	// @Description 证书部署的目标服务器访问配置ID数组，与server_group_id二选一
	// @Example [1,2,3]
	AccessIDs []int `json:"access_ids"`

	// 服务器组ID
	// @Description 证书部署的目标服务器组ID，与access_ids二选一
	// @Example 1
	ServerGroupID int `json:"server_group_id"`

	// 额外配置
	// @Description 工作流额外配置参数，根据部署方式不同而变化
	// @Example {"remote_cert_path":"/etc/nginx/ssl/cert.pem","remote_key_path":"/etc/nginx/ssl/key.pem","pre_command":"nginx -t","post_command":"systemctl reload nginx"}
	ExtraConfig map[string]interface{} `json:"extra_config,omitempty"`

	// 统一通知配置
	// @Description 证书状态变更的统一通知配置，支持多种通知方式
	// @Example {"use_system_global":true,"targets":{"emails":["admin@example.com"]},"events":["apply_success","deploy_failed"]}
	UnifiedNotificationConfig *notification.UnifiedNotificationConfig `json:"unified_notification_config,omitempty"`

	// 自动续期
	// @Description 是否启用证书自动续期功能
	// @Example true
	RenewalEnabled bool `json:"renewal_enabled,omitempty"`

	// 立即执行
	// @Description 是否立即开始执行证书申请流程
	// @Example true
	ExecuteImmediately bool `json:"execute_immediately,omitempty" default:"true"`
}

// @Summary		查看当前用户的所有证书
// @Description	根据jwt携带的id查看当前用户的所有证书
// @Tags			证书管理
// @Security		ApiKeyAuth
// @Accept			json
// @Produce		json
// @Success		200	{object}	response.Response{data=pagination.PaginatedResponse}	"查看证书响应"
// @Failure		400	{string}	string										"无效请求格式"
// @Failure		500	{string}	string										"服务器错误"
// @Router			/cert/certificates [get]
//
// @Description	查看证书请求
func (cc *CertificateController) GetAll(ctx *gin.Context) {
	userID := ctx.GetInt("userID")
	params := pagination.ParsePaginationParams(
		ctx.DefaultQuery("page", "1"),
		ctx.DefaultQuery("pageSize", "10"),
	)

	// 使用新的分页方法获取证书列表和总记录数
	certificates, totalCount, err := cc.client.GetAllWithPagination(ctx.Request.Context(), userID, params.Page, params.PageSize)
	if err != nil {
		logger.Error("GetAllCertificate error", zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 构建分页响应
	page := pagination.NewPaginatedResponse(certificates, totalCount, params.Page, params.PageSize)

	response.Success(ctx, page)
}

// @Summary 创建证书
// @Description 根据域名列表、颁发机构等信息创建新的SSL证书
// @Description
// @Description 支持多种验证方式：
// @Description - DNS验证：通过DNS提供商自动添加TXT记录完成验证
// @Description - HTTP验证：通过在服务器上放置验证文件完成验证
// @Description
// @Description 部署配置：
// @Description - 可通过access_ids指定单个或多个服务器进行部署
// @Description - 可通过server_group_id指定服务器组进行批量部署
// @Description - 通过extra_config传递部署适配器特定配置
// @Description
// @Description 通知配置：
// @Description - 支持Webhook、邮件、短信等多种通知方式
// @Description - 可配置在申请、部署、成功、失败等不同阶段发送通知
// @Tags 证书管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param request body CreateCertificateRequest true "创建证书请求参数"
// @Success 200 {object} response.Response{data=ent.Certificate} "创建成功响应"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /cert/certificates [post]
//
// @Example 基础创建
// @RequestBody {"domains":["example.com","www.example.com"],"issued_by":"letsencrypt","dns_provider_id":1,"validation_type":"dns"}
//
// @Example 带SSH部署
// @RequestBody {"domains":["example.com","www.example.com"],"issued_by":"letsencrypt","dns_provider_id":1,"validation_type":"dns","access_ids":[1],"extra_config":{"remote_cert_path":"/etc/nginx/ssl/cert.pem","remote_key_path":"/etc/nginx/ssl/key.pem","pre_command":"nginx -t","post_command":"systemctl reload nginx"}}
//
// @Example 带tencentcloud-teo部署
// @RequestBody {"domains":["example.com","www.example.com"],"issued_by":"letsencrypt","dns_provider_id":1,"validation_type":"dns","access_ids":[1],"extra_config":{"zoneId":"zone-xxxxxxxxxxxx","endpoint":"teo.intl.tencentcloudapi.com"}}
//
// @Example 带通知配置
// @RequestBody {"domains":["example.com","www.example.com"],"issued_by":"letsencrypt","dns_provider_id":1,"validation_type":"dns","unified_notification_config":{"use_system_global":true,"targets":{"emails":["admin@example.com"]},"events":["apply_success","deploy_failed"]}}
func (cc *CertificateController) Create(ctx *gin.Context) {
	var req CreateCertificateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	uid := ctx.GetInt("userID")

	// 验证部署参数合法性
	if len(req.AccessIDs) > 0 && req.ServerGroupID > 0 {
		response.Error(ctx, http.StatusBadRequest, "accessid和servergroupid不能同时指定")
		return
	}

	// 验证域名列表
	if len(req.Domains) == 0 {
		response.Error(ctx, http.StatusBadRequest, "至少需要提供一个域名")
		return
	}

	// 验证域名格式并去重
	uniqueDomains := make([]string, 0)
	domainMap := make(map[string]bool)
	for _, domain := range req.Domains {
		// 基本域名格式验证
		if domain == "" {
			response.Error(ctx, http.StatusBadRequest, "域名不能为空")
			return
		}

		// 域名格式验证
		if !isValidDomain(domain) {
			response.Error(ctx, http.StatusBadRequest, fmt.Sprintf("无效的域名格式: %s", domain))
			return
		}

		// 去重
		if !domainMap[domain] {
			domainMap[domain] = true
			uniqueDomains = append(uniqueDomains, domain)
		}
	}

	// 检查域名数量限制
	if len(uniqueDomains) > 100 {
		response.Error(ctx, http.StatusBadRequest, "单个证书最多支持100个域名")
		return
	}

	// 记录请求信息
	logger.Info("创建证书请求",
		zap.Int("userID", uid),
		zap.Strings("domains", uniqueDomains),
		zap.String("issuedBy", req.IssuedBy),
		zap.Int("dnsProviderID", req.DNSProviderID),
		zap.String("validationType", req.ValidationType))

	serviceRequest := workflow.CertificateRequest{
		Domains:           uniqueDomains,
		IssuedBy:          req.IssuedBy,
		UserID:            uid,
		DNSProviderID:     req.DNSProviderID,
		ValidationType:    req.ValidationType,
		RenewalDaysBefore: req.RenewalDaysBefore,
		AccessIDs:         req.AccessIDs,
		ServerGroupID:     req.ServerGroupID,
		// 通知配置通过统一配置提供
		UnifiedNotificationConfig: req.UnifiedNotificationConfig,
		// 额外配置
		ExtraConfig: req.ExtraConfig,
		// 续期和执行配置
		RenewalEnabled:     req.RenewalEnabled,
		ExecuteImmediately: req.ExecuteImmediately,
	}

	cert, err := cc.client.CreateCertificate(ctx.Request.Context(), serviceRequest)
	if err != nil {
		logger.Error("CreateCertificate error",
			zap.Error(err),
			zap.Int("userID", uid),
			zap.Strings("domains", uniqueDomains))
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Info("证书创建成功",
		zap.Int("certificateID", cert.ID),
		zap.Int("userID", uid),
		zap.Strings("domains", uniqueDomains))

	response.Success(ctx, cert)
}

// isValidDomain 验证域名格式
func isValidDomain(domain string) bool {
	if domain == "" || len(domain) > 253 {
		return false
	}

	// 域名不能以点开头或结尾
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	// 基本域名格式验证
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?))*$`)
	return domainRegex.MatchString(domain)
}

// getDeployTarget 获取部署目标描述
func getDeployTarget(accessID, serverGroupID int) string {
	if accessID > 0 {
		return fmt.Sprintf("access:%d", accessID)
	} else if serverGroupID > 0 {
		return fmt.Sprintf("servergroup:%d", serverGroupID)
	}
	return "none"
}

//	@Summary		获取证书详情
//	@Description	根据证书ID获取证书详细信息
//	@Tags			证书管理
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int				true	"证书ID"
//	@Success		200	{object}	ent.Certificate	"证书详情"
//	@Failure		400	{string}	string			"无效ID格式"
//	@Failure		404	{string}	string			"证书不存在"
//	@Failure		500	{string}	string			"服务器错误"
//	@Router			/cert/certificates/{id} [get]

func (cc *CertificateController) Get(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid id")
		return
	}

	//service := certificate.NewCertificateService(cc.client)
	cert, err := cc.client.GetByID(ctx.Request.Context(), ctx.GetInt("userID"), id)
	if err != nil {
		if ent.IsNotFound(err) {
			response.Error(ctx, http.StatusNotFound, "certificate not found")
		} else {
			response.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.Success(ctx, cert)
}

// @Summary		删除证书
// @Description	根据证书ID删除指定证书
// @Tags			证书管理
// @Security		ApiKeyAuth
// @Accept			json
// @Produce		json
// @Param			id	path		int					true	"证书ID"
// @Success		200	{object}	response.Response	"删除成功提示"
// @Failure		400	{string}	string				"无效ID格式"
// @Failure		500	{string}	string				"服务器错误"
// @Router			/cert/certificates/{id} [delete]
func (cc *CertificateController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid id")
		return
	}

	//service := certificate.NewCertificateService(cc.client)
	if err := cc.client.Delete(ctx.Request.Context(), ctx.GetInt("userID"), id); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(ctx, nil)
}

// @Summary		下载证书
// @Description	下载指定证书的所有格式（PEM、DER、PFX、JKS等）打包为zip文件
// @Description	支持多种证书格式导出，包括：
// @Description	- PEM格式：包含证书和私钥的文本格式
// @Description	- DER格式：二进制格式的证书
// @Description	- PFX格式：包含证书和私钥的PKCS12格式
// @Description	- JKS格式：Java密钥库格式
// @Tags			证书管理
// @Security		ApiKeyAuth
// @Accept			json
// @Produce		application/zip
// @Param			id	path		int	true	"证书ID"
// @Param			password	query	string	false	"PFX/JKS格式密码（默认：changeit）"
// @Success		200	{file}	binary	"证书压缩包"
// @Failure		400	{string}	string	"无效ID格式"
// @Failure		404	{string}	string	"证书不存在"
// @Failure		500	{string}	string	"服务器错误"
// @Router			/cert/certificates/{id}/download [get]
func (cc *CertificateController) Download(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid id")
		return
	}

	userID := ctx.GetInt("userID")
	password := ctx.DefaultQuery("password", "changeit")

	// 获取证书内容
	certContent, privateKeyContent, err := cc.client.GetCertificateContent(ctx.Request.Context(), userID, id)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, err.Error())
		return
	}

	// 使用certificateutils包中的函数创建证书包
	zipData, err := certificateutils.PackageCertificate(certContent, privateKeyContent, password)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 设置下载头
	ctx.Header("Content-Type", "application/zip")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=certificate_%d_%s.zip", id, time.Now().Format("20060102150405")))

	// 直接写入响应
	ctx.Data(http.StatusOK, "application/zip", zipData)
}
