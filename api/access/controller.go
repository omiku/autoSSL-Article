package access

import (
	"autoSSL/api/response"
	"autoSSL/ent"
	"autoSSL/service/access"
	"autoSSL/utils/pagination"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AccessController 访问配置API控制器
type AccessController struct {
	accessService *access.AccessService
}

// NewAccessController 创建AccessController实例
func NewAccessController(accessService *access.AccessService) *AccessController {
	return &AccessController{accessService: accessService}
}

// CreateAccessRequest 创建访问配置请求参数
//
//	@Schema(title="创建访问配置请求")
type CreateAccessRequest struct {
	Name     string                 `json:"name" binding:"required"`                                // 配置名称
	Type     string                 `json:"type" binding:"required,oneof=ssh api tencentcloud-teo"` // 访问类型
	Config   map[string]interface{} `json:"config" binding:"required"`                              // 配置详情
	IsActive bool                   `json:"is_active,omitempty" default:"true"`                     // 是否启用
}

// UpdateAccessRequest 更新访问配置请求参数
//
//	@Schema(title="更新访问配置请求")
type UpdateAccessRequest struct {
	Name     string                 `json:"name" binding:"required"`                                // 配置名称
	Type     string                 `json:"type" binding:"required,oneof=ssh api tencentcloud-teo"` // 访问类型
	Config   map[string]interface{} `json:"config" binding:"required"`                              // 配置详情
	IsActive bool                   `json:"is_active,omitempty"`                                    // 是否启用
}

// Create 创建访问配置
//
//	@Summary		创建访问配置
//	@Description	创建新的访问配置记录
//	@Tags			access
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			data	body		CreateAccessRequest	true	"访问配置信息"
//	@Success		200		{object}	response.Response{data=ent.Access}
//	@Failure		400		{object}	response.Response
//	@Router			/access [post]
//	@Example		{json} SSH配置示例
//	{
//	  "name": "生产服务器SSH配置",
//	  "type": "ssh",
//	  "config": {
//	    "ssh_host": "192.168.1.100",
//	    "ssh_port": 22,
//	    "ssh_username": "root",
//	    "ssh_password": "your_password",
//	    "ssh_key": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----",
//	    "ssh_key_passphrase": "",
//	    "ssh_auth_method": "password"
//	  },
//	  "is_active": true
//	}
func (c *AccessController) Create(ctx *gin.Context) {
	var req CreateAccessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数验证失败: "+err.Error())
		return
	}

	userID := ctx.GetInt("userID")

	accessConfig, err := c.accessService.Create(
		ctx,
		userID,
		req.Name,
		req.Config,
		req.Type,
		req.IsActive,
	)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "创建访问配置失败: "+err.Error())
		return
	}

	response.Success(ctx, accessConfig)
}

// Get 获取访问配置详情
//
//	@Summary		获取访问配置详情
//	@Description	根据ID获取访问配置详细信息
//	@Tags			access
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"访问配置ID"
//	@Success		200	{object}	response.Response{data=ent.Access}
//	@Failure		404	{object}	response.Response
//	@Router			/access/{id} [get]
func (c *AccessController) Get(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID格式")
		return
	}

	userID := ctx.GetInt("userID")
	// if !exists {
	// 	response.Error(ctx, http.StatusUnauthorized, "未授权访问")
	// 	return
	// }

	accessConfig, err := c.accessService.GetByID(ctx, userID, id)
	if err != nil {
		if ent.IsNotFound(err) {
			response.Error(ctx, http.StatusNotFound, "访问配置不存在")
		} else {
			response.Error(ctx, http.StatusInternalServerError, "获取访问配置失败: "+err.Error())
		}
		return
	}

	response.Success(ctx, accessConfig)
}

// List 获取访问配置列表
//
//	@Summary		获取访问配置列表
//	@Description	根据用户ID获取访问配置列表，支持分页
//	@Tags			access
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int	false	"页码，默认1"
//	@Param			page_size	query		int	false	"每页条数，默认10"
//	@Success		200			{object}	response.Response{data=[]ent.Access}
//	@Failure		500			{object}	response.Response
//	@Router			/access [get]
func (c *AccessController) List(ctx *gin.Context) {
	params := pagination.ParsePaginationParams(
		ctx.DefaultQuery("page", "1"),
		ctx.DefaultQuery("pageSize", "10"),
	)

	userID := ctx.GetInt("userID")

	list, err := c.accessService.ListByUserID(ctx, userID, params.Page, params.PageSize)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取访问配置列表失败: "+err.Error())
		return
	}

	response.Success(ctx, list)
}

// Update 更新访问配置
//
//	@Summary		更新访问配置
//	@Description	根据ID更新访问配置信息
//	@Tags			access
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"访问配置ID"
//	@Param			data	body		UpdateAccessRequest	true	"访问配置信息"
//	@Success		200		{object}	response.Response{data=ent.Access}
//	@Failure		400		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Router			/access/{id} [put]
func (c *AccessController) Update(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID格式")
		return
	}

	var req UpdateAccessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数验证失败: "+err.Error())
		return
	}

	userID := ctx.GetInt("userID")

	accessConfig, err := c.accessService.Update(
		ctx,
		id,
		userID,
		req.Name,
		req.Config,
		req.Type,
		req.IsActive,
	)
	if err != nil {
		if ent.IsNotFound(err) {
			response.Error(ctx, http.StatusNotFound, "访问配置不存在")
		} else {
			response.Error(ctx, http.StatusInternalServerError, "更新访问配置失败: "+err.Error())
		}
		return
	}

	response.Success(ctx, accessConfig)
}

// Delete 删除访问配置
//
//	@Summary		删除访问配置
//	@Description	根据ID删除访问配置
//	@Tags			access
//	@Security		ApiKeyAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"访问配置ID"
//	@Success		200	{object}	response.Response{data=string}
//	@Failure		404	{object}	response.Response
//	@Router			/access/{id} [delete]
func (c *AccessController) Delete(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID格式")
		return
	}

	userID := ctx.GetInt("userID")

	if err := c.accessService.Delete(ctx, id, userID); err != nil {
		if ent.IsNotFound(err) {
			response.Error(ctx, http.StatusNotFound, "访问配置不存在")
		} else {
			response.Error(ctx, http.StatusInternalServerError, "删除访问配置失败: "+err.Error())
		}
		return
	}

	response.Success(ctx, "访问配置删除成功")
}
