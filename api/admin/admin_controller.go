package admin

import (
	"autoSSL/api/response"
	"autoSSL/service/auth"
	"autoSSL/service/global_notification"
	"autoSSL/service/rbac"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	userService               *auth.Service
	rbacService               *rbac.Service
	globalNotificationService *global_notification.Service
}

func NewAdminController(userService *auth.Service, rbacService *rbac.Service, globalNotificationService *global_notification.Service) *AdminController {
	return &AdminController{
		userService:               userService,
		rbacService:               rbacService,
		globalNotificationService: globalNotificationService,
	}
}

// GetAllUsers 获取所有用户
//
//	@Summary		获取所有用户
//	@Description	管理员获取系统所有用户列表（需要管理员权限）
//	@Tags			管理员
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=[]ent.User}	"成功响应"
//	@Failure		500	{string}	string							"服务器错误"
//	@Router			/admin/users [get]
func (c *AdminController) GetAllUsers(ctx *gin.Context) {
	users, err := c.userService.GetAllUsers(ctx)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取用户列表失败")
		return
	}

	response.Success(ctx, users)
}

// UpdateUserRole 更新用户角色
//
//	@Summary		更新用户角色
//	@Description	管理员修改指定用户的角色（需要管理员权限）
//	@Tags			管理员
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response	"成功响应"
//	@Failure		400	{string}	string			"无效请求格式"
//	@Failure		500	{string}	string			"服务器错误"
//	@Router			/admin/users/role [put]
func (c *AdminController) UpdateUserRole(ctx *gin.Context) {
	var req struct {
		UserID int    `json:"user_id" binding:"required"`
		Role   string `json:"role" binding:"required,oneof=admin user"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效请求格式")
		return
	}

	err := c.userService.UpdateUserRole(ctx, req.UserID, req.Role)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "更新用户角色失败")
		return
	}

	response.Success(ctx, nil)
}

// Enforce 检查用户是否有权限
// func (c *AdminController) Enforce(ctx *gin.Context) {
// 	userID, _ := strconv.Atoi(ctx.GetString("userID"))
// 	resource := ctx.Query("resource")
// 	action := ctx.Query("action")

// 	allowed, err := c.rbacService.Enforce(ctx.Request.Context(), userID, resource, action)
// 	if err != nil {
// 		response.Error(ctx, http.StatusInternalServerError, err.Error())
// 		return
// 	}

// 	response.Success(ctx, allowed)
// }

// GetUserRoles 获取用户角色
func (c *AdminController) GetUserRoles(ctx *gin.Context) {
	userID, _ := strconv.Atoi(ctx.Param("userID"))

	roles, err := c.rbacService.GetUserRoles(ctx.Request.Context(), userID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, roles)
}

// GetUserPermissions 获取用户权限
func (c *AdminController) GetUserPermissions(ctx *gin.Context) {
	userID, _ := strconv.Atoi(ctx.Param("userID"))

	permissions, err := c.rbacService.GetUserPermissions(ctx.Request.Context(), userID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, permissions)
}

// AddRoleForUser 为用户添加角色
func (c *AdminController) AddRoleForUser(ctx *gin.Context) {
	type Request struct {
		UserID   int    `json:"user_id" binding:"required"`
		RoleName string `json:"role_name" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.rbacService.AddRoleForUser(ctx.Request.Context(), req.UserID, req.RoleName); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, nil)
}

// RemoveRoleFromUser 从用户移除角色
func (c *AdminController) RemoveRoleFromUser(ctx *gin.Context) {
	type Request struct {
		UserID   int    `json:"user_id" binding:"required"`
		RoleName string `json:"role_name" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.rbacService.RemoveRoleFromUser(ctx.Request.Context(), req.UserID, req.RoleName); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, nil)
}

// CreateRole 创建角色
func (c *AdminController) CreateRole(ctx *gin.Context) {
	type Request struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	r, err := c.rbacService.CreateRole(ctx.Request.Context(), req.Name, req.Description)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, r)
}

// CreatePermission 创建权限
func (c *AdminController) CreatePermission(ctx *gin.Context) {
	type Request struct {
		Resource    string `json:"resource" binding:"required"`
		Action      string `json:"action" binding:"required"`
		Description string `json:"description"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	p, err := c.rbacService.CreatePermission(ctx.Request.Context(), req.Resource, req.Action, req.Description)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, p)
}

// AssignPermissionToRole 为角色分配权限
func (c *AdminController) AssignPermissionToRole(ctx *gin.Context) {
	type Request struct {
		RoleName string `json:"role_name" binding:"required"`
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.rbacService.AssignPermissionToRole(ctx.Request.Context(), req.RoleName, req.Resource, req.Action); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, nil)
}

// RemovePermissionFromRole 从角色移除权限
func (c *AdminController) RemovePermissionFromRole(ctx *gin.Context) {
	type Request struct {
		RoleName string `json:"role_name" binding:"required"`
		Resource string `json:"resource" binding:"required"`
		Action   string `json:"action" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.rbacService.RemovePermissionFromRole(ctx.Request.Context(), req.RoleName, req.Resource, req.Action); err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, nil)
}

// GetAllRoles 获取所有角色
func (c *AdminController) GetAllRoles(ctx *gin.Context) {
	roles, err := c.rbacService.GetAllRoles(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, roles)
}

// GetAllPermissions 获取所有权限
func (c *AdminController) GetAllPermissions(ctx *gin.Context) {
	permissions, err := c.rbacService.GetAllPermissions(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, permissions)
}
