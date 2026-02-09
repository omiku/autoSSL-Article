package dnsprovider

import (
	"autoSSL/api/response"
	"autoSSL/logger"
	dnspd "autoSSL/service/dnsprovider"
	"autoSSL/utils/pagination"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DNSProviderController DNS提供商管理控制器
//
//	@Tags	DNS提供商管理
type DNSProviderController struct {
	service *dnspd.Service
}

func NewDNSProviderController(service *dnspd.Service) *DNSProviderController {
	return &DNSProviderController{service: service}
}

// DNSConfig DNS提供商配置结构体
type DNSConfig struct {
	APIKey    string `json:"api_key,omitempty"`
	APISecret string `json:"api_secret,omitempty"`
	APIToken  string `json:"api_token,omitempty"`
	Region    string `json:"region,omitempty"`
	ZoneID    string `json:"zone_id,omitempty"`
}

// @Schema(description="创建DNS提供商请求参数")
type CreateDNSProviderRequest struct {
	ProviderType string    `json:"provider_type" binding:"required"`
	Config       DNSConfig `json:"config" binding:"required"`
}

// @Schema(description="更新DNS提供商请求参数")
type UpdateDNSProviderRequest struct {
	ProviderType string    `json:"provider_type" binding:"required"`
	Config       DNSConfig `json:"config" binding:"required"`
}

// Validate 根据提供商类型验证配置
func (c *DNSConfig) Validate(providerType string) error {
	switch providerType {
	case "cloudflare":
		// Cloudflare支持API Token或API Key+Email
		if c.APIToken != "" {
			return nil
		}
		if c.APIKey != "" && c.APISecret != "" {
			return nil
		}
		return fmt.Errorf("cloudflare需要api_token或api_key+api_secret")
	case "aliyun", "alidns":
		// 阿里云需要AccessKey ID和AccessKey Secret
		if c.APIKey == "" || c.APISecret == "" {
			return fmt.Errorf("阿里云需要api_key和api_secret")
		}
		return nil
	case "tencentcloud":
		// 腾讯云需要SecretId和SecretKey
		if c.APIKey == "" || c.APISecret == "" {
			return fmt.Errorf("腾讯云需要api_key和api_secret")
		}
		return nil
	default:
		return fmt.Errorf("不支持的提供商类型: %s", providerType)
	}
}

// validateDNSProviderFields 验证供应商类型和配置
func validateDNSProviderFields(providerType string, input CreateDNSProviderRequest) error {
	return input.Config.Validate(providerType)
}

// mapToDNSConfig 将map[string]interface{}转换为DNSConfig结构体
func mapToDNSConfig(configMap map[string]interface{}) DNSConfig {
	config := DNSConfig{}
	if configMap == nil {
		return config
	}

	if val, ok := configMap["api_key"]; ok {
		config.APIKey = fmt.Sprintf("%v", val)
	}
	if val, ok := configMap["api_secret"]; ok {
		config.APISecret = fmt.Sprintf("%v", val)
	}
	if val, ok := configMap["api_token"]; ok {
		config.APIToken = fmt.Sprintf("%v", val)
	}
	if val, ok := configMap["region"]; ok {
		config.Region = fmt.Sprintf("%v", val)
	}
	if val, ok := configMap["zone_id"]; ok {
		config.ZoneID = fmt.Sprintf("%v", val)
	}
	return config
}

// @Summary		创建DNS提供商配置
// @Description	根据提供商类型、API密钥等信息创建新的DNS提供商配置
// @Tags			DNS提供商管理
// @Security		ApiKeyAuth
// @Accept			json
// @Produce		json
// @Param			request	body		CreateDNSProviderRequest	true	"创建DNS提供商请求参数"
// @Success		200		{object}	response.Response			"创建成功响应"
// @Failure		400		{string}	string						"无效请求格式"
// @Failure		500		{string}	string						"服务器错误"
// @Router			/dns-provider/dnsproviders [post]
func (dpc *DNSProviderController) Create(ctx *gin.Context) {
	var input CreateDNSProviderRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// 验证字段：根据供应商类型检查所需字段
	if err := validateDNSProviderFields(input.ProviderType, input); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// 获取当前用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "未授权访问")
		return
	}

	// 将DNSConfig转换为map
	configMap := map[string]interface{}{
		"api_key":    input.Config.APIKey,
		"api_secret": input.Config.APISecret,
		"api_token":  input.Config.APIToken,
		"region":     input.Config.Region,
		"zone_id":    input.Config.ZoneID,
	}

	// 移除空值
	for k, v := range configMap {
		if v == "" {
			delete(configMap, k)
		}
	}

	// 在CreateDNSProvider方法中更新服务调用
	provider, err := dpc.service.Create(ctx.Request.Context(), userID.(int), input.ProviderType, configMap)

	if err != nil {
		logger.Debug("create dns provider failed", zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, provider)
}

// @Summary		获取DNS提供商列表
// @Description	获取所有DNS提供商的列表，支持分页查询
// @Tags			DNS提供商管理
// @Security		ApiKeyAuth
// @Accept			json
// @Produce		json
// @Param			page		query		int		false	"页码，默认为1"
// @Param			pageSize	query		int		false	"每页数量，默认为10"
// @Success		200	{object}	response.Response{data=pagination.PaginatedResponse}	"DNS提供商分页列表"
// @Failure		500	{string}	string									"服务器错误"
// @Router			/dns-provider/dnsproviders [get]
func (dpc *DNSProviderController) GetAll(ctx *gin.Context) {
	params := pagination.ParsePaginationParams(
		ctx.DefaultQuery("page", "1"),
		ctx.DefaultQuery("pageSize", "10"),
	)
	providers, total, err := dpc.service.GetAll(ctx.Request.Context(), ctx.GetInt("userID"), params.Page, params.PageSize)
	if err != nil {
		logger.Error("get dns providers failed", zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// 转换providers为包含DNSConfig的响应
	var responseProviders []gin.H
	for _, provider := range providers {
		config := mapToDNSConfig(provider.Config)

		responseProviders = append(responseProviders, gin.H{
			"id":            provider.ID,
			"provider_type": provider.ProviderType,
			"config":        config,
			"created_at":    provider.CreatedAt,
			"updated_at":    provider.UpdatedAt,
		})
	}

	response.Success(ctx, pagination.NewPaginatedResponse(responseProviders, total, params.Page, params.PageSize))
}

// @Summary		获取DNS提供商详情
// @Description	根据DNS提供商ID获取详细配置信息
// @Tags			DNS提供商管理
// @Security		ApiKeyAuth
// @Accept			json
// @Produce		json
// @Param			id	path		int					true	"DNS提供商ID"
// @Success		200	{object}	response.Response	"DNS提供商详情"
// @Failure		400	{string}	string				"无效ID格式"
// @Failure		500	{string}	string				"服务器错误"
// @Router			/dns-provider/dnsproviders/{id} [get]
func (dpc *DNSProviderController) Get(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.Error(ctx, http.StatusBadRequest, "id is required")
		return
	}

	// 将从url中获取的id转为int类型
	intid, err := strconv.Atoi(id)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "id is required")
	}

	// todo 这里需要对用户进行校验，确保该dns供应商是否属于该用户
	userID := ctx.GetInt("userID")

	provider, err := dpc.service.GetByID(ctx.Request.Context(), userID, intid)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Println(provider)

	// 将map转换为DNSConfig结构体
	config := mapToDNSConfig(provider.Config)

	response.Success(ctx, gin.H{
		"id":            provider.ID,
		"provider_type": provider.ProviderType,
		"config":        config,
		"created_at":    provider.CreatedAt,
		"updated_at":    provider.UpdatedAt,
	})
}

// @Summary		更新DNS提供商配置
// @Description	根据ID和新参数更新现有DNS提供商配置
// @Tags			DNS提供商管理
// @Security		ApiKeyAuth
// @Accept			json
// @Produce		json
// @Param			id		path		int										true	"DNS提供商ID"
// @Param			request	body		dnsprovider.UpdateDNSProviderRequest	true	"更新DNS提供商请求参数"
// @Success		200		{object}	response.Response						"更新成功响应"
// @Failure		400		{string}	string									"无效ID或请求格式"
// @Failure		500		{string}	string									"服务器错误"
// @Router			/dns-provider/dnsproviders/{id} [put]
func (dpc *DNSProviderController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.Error(ctx, http.StatusBadRequest, "id is required")
		return
	}

	var input UpdateDNSProviderRequest

	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// 验证字段：根据供应商类型检查所需字段
	if err := validateDNSProviderFields(input.ProviderType, CreateDNSProviderRequest(input)); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// 将从url中获取的id转为int类型
	intid, err := strconv.Atoi(id)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "id is required")
		return
	}

	userID := ctx.GetInt("userID")

	// 将DNSConfig转换为map
	configMap := map[string]interface{}{
		"api_key":    input.Config.APIKey,
		"api_secret": input.Config.APISecret,
		"api_token":  input.Config.APIToken,
		"region":     input.Config.Region,
		"zone_id":    input.Config.ZoneID,
	}

	// 移除空值
	for k, v := range configMap {
		if v == "" {
			delete(configMap, k)
		}
	}

	// 在UpdateDNSProvider方法中更新服务调用
	provider, err := dpc.service.Update(ctx.Request.Context(), userID, intid, input.ProviderType, configMap)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, provider)
}

// @Summary		删除DNS提供商配置
// @Description	根据ID删除指定的DNS提供商配置
// @Tags			DNS提供商管理
// @Security		ApiKeyAuth
// @Accept			json
// @Produce		json
// @Param			id	path		int					true	"DNS提供商ID"
// @Success		200	{object}	response.Response	"删除成功提示"
// @Failure		400	{string}	string				"无效ID格式"
// @Failure		500	{string}	string				"服务器错误"
// @Router			/dns-provider/dnsproviders/{id} [delete]
func (dpc *DNSProviderController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		response.Error(ctx, http.StatusBadRequest, "id is required")
		return
	}

	intid, err := strconv.Atoi(id)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "id is required")
		return
	}

	userID := ctx.GetInt("userID")
	if err := dpc.service.Delete(ctx.Request.Context(), userID, intid); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, nil)
}
