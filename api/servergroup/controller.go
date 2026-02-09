package servergroup

import (
	"autoSSL/api/response"
	"autoSSL/service/servergroup"
	"autoSSL/utils/pagination"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ServerGroupController 服务器组控制器
// @Schema 服务器组管理控制器
// @Description 提供服务器组的创建、查询、更新、删除等RESTful API接口
// @Tag 服务器组管理
// @BasePath /api/v1
// @Produces json
// @Consumes json
type ServerGroupController struct {
	service *servergroup.ServerGroupService
}

// NewServerGroupController 创建服务器组控制器实例
func NewServerGroupController(service *servergroup.ServerGroupService) *ServerGroupController {
	return &ServerGroupController{service: service}
}

// Create 创建服务器组
// @Summary 创建服务器组
// @Description 创建一个新的服务器组，并关联指定的访问权限
// @Tags 服务器组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "token"
// @Param request body servergroup.CreateServerGroupRequest true "服务器组创建请求"
// @Success 200 {object} response.Response{data=ent.ServerGroup} "创建成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /servers [post]
// @Example 请求示例:
//
//	{
//	  "name": "生产环境服务器组",
//	  "description": "生产环境所有服务器",
//	  "status": "active",
//	  "access_ids": [
//	    {"access_id": 1, "type": "ssh"},
//	    {"access_id": 2, "type": "ssh"}
//	  ]
//	}
func (c *ServerGroupController) Create(ctx *gin.Context) {
	var req servergroup.CreateServerGroupRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效请求格式")
		return
	}

	userID := ctx.GetInt("userID")
	req.UserID = userID
	serverGroup, err := c.service.Create(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "不属于当前用户") {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else if strings.Contains(err.Error(), "类型必须一致") {
			response.Error(ctx, http.StatusBadRequest, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "创建服务器组失败")
		}
		return
	}

	response.Success(ctx, serverGroup)
}

// Get 获取服务器组列表及关联Access数量
// @Summary 获取服务器组列表
// @Description 获取当前用户的服务器组列表，包含每个服务器组关联的访问权限数量
// @Tags 服务器组管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页条数，默认10"
// @Success 200 {object} response.Response{data=[]ent.ServerGroup} "服务器组列表"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /servers [get]
func (c *ServerGroupController) Get(ctx *gin.Context) {
	params := pagination.ParsePaginationParams(
		ctx.DefaultQuery("page", "1"),
		ctx.DefaultQuery("pageSize", "10"),
	)

	userID := ctx.GetInt("userID")
	servers, err := c.service.ListWithAccessCount(ctx, userID, params.Page, params.PageSize)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取服务器组列表失败")
		return
	}

	response.Success(ctx, servers)
}

// Update 更新服务器组信息
// @Summary 更新服务器组
// @Description 更新指定服务器组的信息，包括名称、描述、状态和关联的访问权限
// @Tags 服务器组管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "服务器组ID"
// @Param request body servergroup.UpdateServerGroupRequest true "服务器组更新请求"
// @Success 200 {object} response.Response{data=ent.ServerGroup} "更新成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /servers/{id} [put]
// @Example 请求示例:
//
//	{
//	  "id": 1,
//	  "name": "更新后的服务器组",
//	  "description": "更新后的描述",
//	  "status": "inactive",
//	  "access_ids": [
//	    {"access_id": 1, "type": "ssh"}
//	  ]
//	}
func (c *ServerGroupController) Update(ctx *gin.Context) {
	var req servergroup.UpdateServerGroupRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效请求格式")
		return
	}

	userID := ctx.GetInt("userID")
	req.UserID = userID
	serverGroup, err := c.service.Update(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "不属于当前用户") {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else if strings.Contains(err.Error(), "类型必须一致") {
			response.Error(ctx, http.StatusBadRequest, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "更新服务器组失败")
		}
		return
	}

	response.Success(ctx, serverGroup)
}

// GetByID 获取服务器组详情
// @Summary 获取服务器组详情
// @Description 根据ID获取服务器组的详细信息，包括关联的访问权限
// @Tags 服务器组管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "服务器组ID"
// @Success 200 {object} response.Response{data=ent.ServerGroup} "服务器组详情"
// @Failure 400 {object} response.Response "无效ID格式"
// @Failure 404 {object} response.Response "服务器组不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /servers/{id} [get]
func (c *ServerGroupController) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效ID格式")
		return
	}

	userID := ctx.GetInt("userID")
	serverGroup, err := c.service.GetByID(ctx, userID, id)
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.Error(ctx, http.StatusNotFound, "服务器组不存在")
		} else {
			response.Error(ctx, http.StatusInternalServerError, "获取服务器组详情失败")
		}
		return
	}

	response.Success(ctx, serverGroup)
}

// Delete 删除服务器组
// @Summary 删除服务器组
// @Description 删除指定的服务器组，同时移除所有关联的服务器组成员
// @Tags 服务器组管理
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "服务器组ID"
// @Success 200 {object} response.Response "删除成功"
// @Failure 400 {object} response.Response "无效ID格式"
// @Failure 403 {object} response.Response "权限不足"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /servers/{id} [delete]
// @Example 请求示例:
// DELETE /api/v1/servers/1
func (c *ServerGroupController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效ID格式")
		return
	}

	userID := ctx.GetInt("userID")
	if err := c.service.Delete(ctx, userID, id); err != nil {
		if strings.Contains(err.Error(), "不属于当前用户") {
			response.Error(ctx, http.StatusForbidden, err.Error())
		} else {
			response.Error(ctx, http.StatusInternalServerError, "删除服务器组失败")
		}
		return
	}

	response.Success(ctx, nil)
}
