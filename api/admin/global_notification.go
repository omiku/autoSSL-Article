package admin

import (
	"autoSSL/api/response"
	"autoSSL/service/global_notification"

	"github.com/gin-gonic/gin"
)

// GlobalEmailConfigResponse 全局邮件配置响应结构（管理员）
type GlobalEmailConfigResponse struct {
	SMTPServer   string `json:"smtp_server"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
	IsConfigured bool   `json:"is_configured"`
}

// AdminGlobalSMSConfigResponse 全局短信配置响应结构（管理员）
type AdminGlobalSMSConfigResponse struct {
	Provider     string `json:"provider"`
	AccessKey    string `json:"access_key"`
	SecretKey    string `json:"secret_key"`
	Region       string `json:"region"`
	SignName     string `json:"sign_name"`
	TemplateCode string `json:"template_code"`
	IsConfigured bool   `json:"is_configured"`
}

// SetEmailConfigRequest 设置邮件配置请求结构
type SetEmailConfigRequest struct {
	SMTPServer   string `json:"smtp_server" binding:"required"`
	SMTPPort     int    `json:"smtp_port" binding:"required"`
	SMTPUsername string `json:"smtp_username" binding:"required"`
	SMTPPassword string `json:"smtp_password" binding:"required"`
	FromEmail    string `json:"from_email" binding:"required"`
	FromName     string `json:"from_name" binding:"required"`
}

// SetSMSConfigRequest 设置短信配置请求结构
type SetSMSConfigRequest struct {
	Provider     string `json:"provider" binding:"required"`
	AccessKey    string `json:"access_key" binding:"required"`
	SecretKey    string `json:"secret_key" binding:"required"`
	Region       string `json:"region" binding:"required"`
	SignName     string `json:"sign_name" binding:"required"`
	TemplateCode string `json:"template_code" binding:"required"`
}

// AdminGetGlobalEmailConfig 管理员获取全局邮件配置
// @Summary 管理员获取全局邮件配置
// @Description 管理员获取完整的全局邮件配置信息
// @Tags 全局通知配置
// @Produce json
// @Success 200 {object} response.Response{data=GlobalEmailConfigResponse}
// @Failure 500 {object} response.Response
// @Router /admin/global-config/email [get]
func (c *AdminController) AdminGetGlobalEmailConfig(ctx *gin.Context) {
	config, err := c.globalNotificationService.GetGlobalEmailConfig(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, 500, "获取邮件配置失败")
		return
	}

	if config == nil {
		response.Success(ctx, GlobalEmailConfigResponse{
			IsConfigured: false,
		})
		return
	}

	response.Success(ctx, GlobalEmailConfigResponse{
		SMTPServer:   config.SMTPServer,
		SMTPPort:     config.SMTPPort,
		SMTPUsername: config.Username,
		FromEmail:    config.From,
		FromName:     config.FromName,
		IsConfigured: true,
	})
}

// AdminSetGlobalEmailConfig 管理员设置全局邮件配置
// @Summary 管理员设置全局邮件配置
// @Description 管理员设置全局邮件配置信息
// @Tags 全局通知配置
// @Accept json
// @Produce json
// @Param config body SetEmailConfigRequest true "邮件配置信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/global-config/email [post]
func (c *AdminController) AdminSetGlobalEmailConfig(ctx *gin.Context) {
	var req SetEmailConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, 400, "参数错误")
		return
	}

	if err := c.globalNotificationService.SetGlobalEmailConfig(ctx.Request.Context(), &global_notification.GlobalEmailConfig{
		SMTPServer: req.SMTPServer,
		SMTPPort:   req.SMTPPort,
		Username:   req.SMTPUsername,
		Password:   req.SMTPPassword,
		From:       req.FromEmail,
		FromName:   req.FromName,
	}); err != nil {
		response.Error(ctx, 500, "保存邮件配置失败")
		return
	}

	response.Success(ctx, nil)
}

// AdminGetGlobalSMSConfig 管理员获取全局短信配置
// @Summary 管理员获取全局短信配置
// @Description 管理员获取完整的全局短信配置信息
// @Tags 全局通知配置
// @Produce json
// @Success 200 {object} response.Response{data=AdminGlobalSMSConfigResponse}
// @Failure 500 {object} response.Response
// @Router /admin/global-config/sms [get]
func (c *AdminController) AdminGetGlobalSMSConfig(ctx *gin.Context) {
	config, err := c.globalNotificationService.GetGlobalSMSConfig(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, 500, "获取短信配置失败")
		return
	}

	if config == nil {
		response.Success(ctx, AdminGlobalSMSConfigResponse{
			IsConfigured: false,
		})
		return
	}

	response.Success(ctx, AdminGlobalSMSConfigResponse{
		Provider:     "aliyun", // 暂时固定为阿里云
		AccessKey:    config.APIKey,
		SecretKey:    config.APISecret,
		Region:       "cn-hangzhou", // 暂时固定
		SignName:     config.SignName,
		TemplateCode: config.TemplateCode,
		IsConfigured: true,
	})
}

// AdminSetGlobalSMSConfig 管理员设置全局短信配置
// @Summary 管理员设置全局短信配置
// @Description 管理员设置全局短信配置信息
// @Tags 全局通知配置
// @Accept json
// @Produce json
// @Param config body SetSMSConfigRequest true "短信配置信息"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /admin/global-config/sms [post]
func (c *AdminController) AdminSetGlobalSMSConfig(ctx *gin.Context) {
	var req SetSMSConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, 400, "参数错误")
		return
	}

	if err := c.globalNotificationService.SetGlobalSMSConfig(ctx.Request.Context(), &global_notification.GlobalSMSConfig{
		APIKey:       req.AccessKey,
		APISecret:    req.SecretKey,
		SignName:     req.SignName,
		TemplateCode: req.TemplateCode,
	}); err != nil {
		response.Error(ctx, 500, "保存短信配置失败")
		return
	}

	response.Success(ctx, nil)
}
