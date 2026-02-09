package notification

import (
	"fmt"
	"net/http"
	"strconv"

	"autoSSL/api/response"
	"autoSSL/ent"
	"autoSSL/ent/notificationconfig"
	"autoSSL/ent/user"
	"autoSSL/service/global_notification"
	"autoSSL/service/notification"

	"github.com/gin-gonic/gin"
)

// Controller 通知配置控制器
type Controller struct {
	client                    *ent.Client
	globalNotificationService *global_notification.Service
}

// TemplateController 模板管理控制器
type TemplateController struct {
	notificationService *notification.TemplateService
}

// NewController 创建通知配置控制器
func NewController(client *ent.Client, globalNotificationService *global_notification.Service) *Controller {
	return &Controller{
		client:                    client,
		globalNotificationService: globalNotificationService,
	}
}

// NewTemplateController 创建模板控制器
func NewTemplateController(notificationService *notification.TemplateService) *TemplateController {
	return &TemplateController{
		notificationService: notificationService,
	}
}

// ListNotificationConfigs 获取用户的通知配置列表
// @Summary 获取通知配置列表
// @Description 获取当前用户的所有通知配置
// @Tags 通知配置
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} response.Response{data=[]ent.NotificationConfig} "通知配置列表"
// @Failure 500 {object} response.Response "获取通知配置失败"
// @Router /notification/config [get]
func (c *Controller) ListNotificationConfigs(ctx *gin.Context) {
	userID := ctx.GetInt("userID")

	configs, err := c.client.NotificationConfig.Query().
		Where(notificationconfig.HasUserWith(user.IDEQ(userID))).
		All(ctx.Request.Context())

	if err != nil {
		response.Error(ctx, 500, "获取通知配置失败")
		return
	}

	response.Success(ctx, configs)
}

// CreateNotificationConfig 创建通知配置
// @Summary 创建通知配置
// @Description 创建新的通知配置
// @Tags 通知配置
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param config body CreateConfigRequest true "通知配置信息"
// @Success 200 {object} response.Response{data=ent.NotificationConfig} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "创建通知配置失败"
// @Router /notification/config [post]
func (c *Controller) CreateNotificationConfig(ctx *gin.Context) {
	userID := ctx.GetInt("userID")

	var req CreateConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, 400, "参数错误")
		return
	}

	// 获取用户
	user, err := c.client.User.Get(ctx.Request.Context(), userID)
	if err != nil {
		response.Error(ctx, 404, "用户不存在")
		return
	}

	// 转换通知类型
	configType, err := parseNotificationType(req.Type)
	if err != nil {
		response.Error(ctx, 400, err.Error())
		return
	}

	config, err := c.client.NotificationConfig.Create().
		SetName(req.Name).
		SetType(configType).
		SetConfig(req.Config).
		SetIsActive(req.IsActive).
		SetUser(user).
		Save(ctx.Request.Context())

	if err != nil {
		response.Error(ctx, 500, "创建通知配置失败")
		return
	}

	response.Success(ctx, config)
}

// UpdateNotificationConfig 更新通知配置
// @Summary 更新通知配置
// @Description 更新指定的通知配置
// @Tags 通知配置
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "通知配置ID"
// @Param config body UpdateConfigRequest true "更新的通知配置信息"
// @Success 200 {object} response.Response{data=ent.NotificationConfig} "更新成功"
// @Failure 400 {object} response.Response "ID格式错误或请求参数错误"
// @Failure 404 {object} response.Response "通知配置不存在或无权限"
// @Failure 500 {object} response.Response "更新通知配置失败"
// @Router /notification/config/{id} [put]
func (c *Controller) UpdateNotificationConfig(ctx *gin.Context) {
	userID := ctx.GetInt("userID")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, 400, "ID格式错误")
		return
	}

	var req UpdateConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, 400, "参数错误")
		return
	}

	// 查询配置并验证权限
	config, err := c.client.NotificationConfig.Query().
		Where(
			notificationconfig.ID(id),
			notificationconfig.HasUserWith(user.IDEQ(userID)),
		).
		Only(ctx.Request.Context())

	if err != nil {
		response.Error(ctx, 404, "通知配置不存在或无权限")
		return
	}

	// 构建更新操作
	update := config.Update()
	if req.Name != nil {
		update.SetName(*req.Name)
	}
	if req.Type != nil {
		configType, err := parseNotificationType(*req.Type)
		if err != nil {
			response.Error(ctx, 400, err.Error())
			return
		}
		update.SetType(configType)
	}
	if req.Config != nil {
		update.SetConfig(*req.Config)
	}
	if req.IsActive != nil {
		update.SetIsActive(*req.IsActive)
	}

	updatedConfig, err := update.Save(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, 500, "更新通知配置失败")
		return
	}

	response.Success(ctx, updatedConfig)
}

// DeleteNotificationConfig 删除通知配置
// @Summary 删除通知配置
// @Description 删除指定的通知配置
// @Tags 通知配置
// @Security ApiKeyAuth
// @Param id path int true "通知配置ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "ID格式错误"
// @Failure 404 {object} response.Response "通知配置不存在或无权限"
// @Failure 500 {object} response.Response "删除通知配置失败"
// @Router /notification/config/{id} [delete]
func (c *Controller) DeleteNotificationConfig(ctx *gin.Context) {
	userID := ctx.GetInt("userID")
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, 400, "ID格式错误")
		return
	}

	// 查询配置并验证权限
	config, err := c.client.NotificationConfig.Query().
		Where(
			notificationconfig.ID(id),
			notificationconfig.HasUserWith(user.IDEQ(userID)),
		).
		Only(ctx.Request.Context())

	if err != nil {
		response.Error(ctx, 404, "通知配置不存在或无权限")
		return
	}

	if err := c.client.NotificationConfig.DeleteOne(config).Exec(ctx.Request.Context()); err != nil {
		response.Error(ctx, 500, "删除通知配置失败")
		return
	}

	response.Success(ctx, nil)
}

// CreateConfigRequest 创建通知配置请求结构
type CreateConfigRequest struct {
	Name     string                 `json:"name" binding:"required"`
	Type     string                 `json:"type" binding:"required,oneof=webhook email sms"`
	Config   map[string]interface{} `json:"config" binding:"required"`
	IsActive bool                   `json:"is_active"`
}

// UpdateConfigRequest 更新通知配置请求结构
type UpdateConfigRequest struct {
	Name     *string                 `json:"name"`
	Type     *string                 `json:"type" binding:"omitempty,oneof=webhook email sms"`
	Config   *map[string]interface{} `json:"config"`
	IsActive *bool                   `json:"is_active"`
}

// parseNotificationType 将字符串类型转换为ent类型
func parseNotificationType(typeStr string) (notificationconfig.Type, error) {
	switch typeStr {
	case "webhook":
		return notificationconfig.TypeWebhook, nil
	case "email":
		return notificationconfig.TypeEmail, nil
	case "sms":
		return notificationconfig.TypeSms, nil
	default:
		return "", fmt.Errorf("无效的通知类型")
	}
}

// GetGlobalEmailConfig 用户获取全局邮件配置（只读，隐藏敏感信息）
// @Summary 用户获取全局邮件配置
// @Description 用户获取全局邮件配置，只显示是否已配置，不显示敏感信息
// @Tags 通知配置
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} response.Response{data=UserEmailConfigResponse} "全局邮件配置"
// @Failure 500 {object} response.Response "获取邮件配置失败"
// @Router /notification/global/email [get]
func (c *Controller) GetGlobalEmailConfig(ctx *gin.Context) {
	config, err := c.globalNotificationService.GetGlobalEmailConfig(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, 500, "获取邮件配置失败")
		return
	}

	response.Success(ctx, UserEmailConfigResponse{
		FromName:     config.FromName,
		IsConfigured: config != nil && config.IsValid(),
	})
}

// GetGlobalSMSConfig 用户获取全局短信配置（只读，隐藏敏感信息）
// @Summary 用户获取全局短信配置
// @Description 用户获取全局短信配置，只显示是否已配置，不显示敏感信息
// @Tags 通知配置
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} response.Response{data=UserSMSConfigResponse} "全局短信配置"
// @Failure 500 {object} response.Response "获取短信配置失败"
// @Router /notification/global/sms [get]
func (c *Controller) GetGlobalSMSConfig(ctx *gin.Context) {
	config, err := c.globalNotificationService.GetGlobalSMSConfig(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, 500, "获取短信配置失败")
		return
	}

	response.Success(ctx, UserSMSConfigResponse{
		SignName:     config.SignName,
		IsConfigured: config != nil && config.IsValid(),
	})
}

// UserEmailConfigResponse 全局邮件配置响应结构（用户）
type UserEmailConfigResponse struct {
	FromName     string `json:"from_name"`
	IsConfigured bool   `json:"is_configured"`
}

// UserSMSConfigResponse 全局短信配置响应结构（用户）
type UserSMSConfigResponse struct {
	SignName     string `json:"sign_name"`
	IsConfigured bool   `json:"is_configured"`
}

// CreateTemplateRequest 创建模板请求结构
type CreateTemplateRequest struct {
	Name            string `json:"name" binding:"required"`
	TemplateType    string `json:"template_type" binding:"required"`
	SubjectTemplate string `json:"subject_template"`
	ContentTemplate string `json:"content_template" binding:"required"`
	Format          string `json:"format" binding:"required"`
}

// UpdateTemplateRequest 更新模板请求结构
type UpdateTemplateRequest struct {
	Name            string `json:"name" binding:"required"`
	SubjectTemplate string `json:"subject_template"`
	ContentTemplate string `json:"content_template" binding:"required"`
	Format          string `json:"format" binding:"required"`
}

// TemplateResponse 模板响应结构
type TemplateResponse struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	TemplateType    string `json:"template_type"`
	SubjectTemplate string `json:"subject_template"`
	ContentTemplate string `json:"content_template"`
	Format          string `json:"format"`
	IsSystemDefault bool   `json:"is_system_default"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

// CreateTemplate 创建用户模板
// @Summary 创建用户自定义模板
// @Description 创建用户自定义通知模板
// @Tags 通知模板
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param template body CreateTemplateRequest true "模板信息"
// @Success 200 {object} response.Response{data=TemplateResponse} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/templates [post]
func (c *TemplateController) CreateTemplate(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "用户未登录")
		return
	}

	var req CreateTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	template, err := c.notificationService.CreateUserTemplate(
		ctx.Request.Context(),
		int(userID.(float64)),
		req.Name,
		req.TemplateType,
		req.SubjectTemplate,
		req.ContentTemplate,
		req.Format,
	)

	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "创建模板失败: "+err.Error())
		return
	}

	response.Success(ctx, c.convertToResponse(template))
}

// UpdateTemplate 更新用户模板
// @Summary 更新用户自定义模板
// @Description 更新用户自定义通知模板
// @Tags 通知模板
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Param template body UpdateTemplateRequest true "模板信息"
// @Success 200 {object} response.Response{data=TemplateResponse} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "模板不存在或无权限"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/templates/{id} [put]
func (c *TemplateController) UpdateTemplate(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "用户未登录")
		return
	}

	templateID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的模板ID")
		return
	}

	var req UpdateTemplateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	template, err := c.notificationService.UpdateUserTemplate(
		ctx.Request.Context(),
		templateID,
		int(userID.(float64)),
		req.Name,
		req.SubjectTemplate,
		req.ContentTemplate,
		req.Format,
	)

	if err != nil {
		if err.Error() == "模板不存在或无权限" {
			response.Error(ctx, http.StatusForbidden, "模板不存在或无权限")
		} else {
			response.Error(ctx, http.StatusInternalServerError, "更新模板失败: "+err.Error())
		}
		return
	}

	response.Success(ctx, c.convertToResponse(template))
}

// DeleteTemplate 删除用户模板
// @Summary 删除用户自定义模板
// @Description 删除用户自定义通知模板
// @Tags 通知模板
// @Security ApiKeyAuth
// @Param id path int true "模板ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "模板不存在或无权限"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/templates/{id} [delete]
func (c *TemplateController) DeleteTemplate(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "用户未登录")
		return
	}

	templateID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的模板ID")
		return
	}

	err = c.notificationService.DeleteUserTemplate(
		ctx.Request.Context(),
		templateID,
		int(userID.(float64)),
	)

	if err != nil {
		if err.Error() == "模板不存在或无权限" {
			response.Error(ctx, http.StatusForbidden, "模板不存在或无权限")
		} else {
			response.Error(ctx, http.StatusInternalServerError, "删除模板失败: "+err.Error())
		}
		return
	}

	response.Success(ctx, nil)
}

// ListUserTemplates 获取用户模板列表
// @Summary 获取用户自定义模板列表
// @Description 获取当前用户的所有自定义模板
// @Tags 通知模板
// @Security ApiKeyAuth
// @Param type query string false "模板类型 (email, sms, webhook)"
// @Produce json
// @Success 200 {object} response.Response{data=[]TemplateResponse} "模板列表"
// @Failure 401 {object} response.Response "用户未登录"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/templates [get]
func (c *TemplateController) ListUserTemplates(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "用户未登录")
		return
	}

	templateType := ctx.Query("type")

	templates, err := c.notificationService.ListUserTemplates(
		ctx.Request.Context(),
		int(userID.(float64)),
		templateType,
	)

	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取模板列表失败: "+err.Error())
		return
	}

	var responses []TemplateResponse
	for _, template := range templates {
		responses = append(responses, *c.convertToResponse(template))
	}

	response.Success(ctx, responses)
}

// ListSystemTemplates 获取系统模板列表
// @Summary 获取系统默认模板列表
// @Description 获取所有系统默认模板
// @Tags 通知模板
// @Security ApiKeyAuth
// @Param type query string false "模板类型 (email, sms, webhook)"
// @Produce json
// @Success 200 {object} response.Response{data=[]TemplateResponse} "系统模板列表"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/templates/system [get]
func (c *TemplateController) ListSystemTemplates(ctx *gin.Context) {
	templateType := ctx.Query("type")

	templates, err := c.notificationService.ListAllTemplates(ctx.Request.Context(), 0, templateType)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取系统模板失败: "+err.Error())
		return
	}

	// 过滤系统模板
	var systemTemplates []*ent.NotificationTemplate
	for _, template := range templates {
		if template.IsSystemDefault {
			systemTemplates = append(systemTemplates, template)
		}
	}

	var responses []TemplateResponse
	for _, template := range systemTemplates {
		responses = append(responses, *c.convertToResponse(template))
	}

	response.Success(ctx, responses)
}

// GetTemplate 获取单个模板详情
// @Summary 获取模板详情
// @Description 获取指定模板的详细信息
// @Tags 通知模板
// @Security ApiKeyAuth
// @Param id path int true "模板ID"
// @Produce json
// @Success 200 {object} response.Response{data=TemplateResponse} "模板详情"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "模板不存在或无权限"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/templates/{id} [get]
func (c *TemplateController) GetTemplate(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "用户未登录")
		return
	}

	templateID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的模板ID")
		return
	}

	// 获取所有可用模板（包括用户和系统模板）
	allTemplates, err := c.notificationService.ListAllTemplates(
		ctx.Request.Context(),
		int(userID.(float64)),
		"",
	)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取模板失败: "+err.Error())
		return
	}

	// 查找指定模板
	for _, template := range allTemplates {
		if template.ID == templateID {
			response.Success(ctx, c.convertToResponse(template))
			return
		}
	}

	response.Error(ctx, http.StatusNotFound, "模板不存在")
}

// GetTemplateOptions 获取模板类型选项
// @Summary 获取通知模板类型选项
// @Description 获取可用的通知模板类型列表
// @Tags 通知模板
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} response.Response{data=map[string][]string} "模板类型列表"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/templates/options [get]
func (c *TemplateController) TemplateTypeOptions(ctx *gin.Context) {
	options := map[string][]string{
		"template_types": {"email", "sms", "webhook"},
		"event_types": {
			"apply_success",
			"apply_failed",
			"deploy_success",
			"deploy_failed",
			"all_success",
			"all_failed",
			"step_started",
			"workflow_failed",
			"renewal_reminder",
		},
		"formats": {"html", "text", "json"},
	}
	response.Success(ctx, options)
}

// convertToResponse 转换模板为响应格式
func (c *TemplateController) convertToResponse(template *ent.NotificationTemplate) *TemplateResponse {
	return &TemplateResponse{
		ID:              template.ID,
		Name:            template.Name,
		TemplateType:    template.Type,
		SubjectTemplate: template.SubjectTemplate,
		ContentTemplate: template.ContentTemplate,
		Format:          template.Format,
		IsSystemDefault: template.IsSystemDefault,
		CreatedAt:       template.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       template.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
