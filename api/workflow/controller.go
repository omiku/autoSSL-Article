package workflow

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"autoSSL/api/response"
	"autoSSL/service/workflow"
	"autoSSL/utils/pagination"
)

// Controller 工作流API控制器
// @Schema 工作流管理控制器
// @Description 提供工作流的查询、管理、控制等RESTful API接口
// @Tag 工作流管理
// @BasePath /api/v1
// @Produces json
// @Consumes json
type Controller struct {
	workflowMgr *workflow.Manager
}

// NewController 创建工作流控制器实例
func NewController(workflowMgr *workflow.Manager) *Controller {
	return &Controller{workflowMgr: workflowMgr}
}

// GetWorkflowByID 获取工作流详情
// @Summary 获取工作流详情
// @Description 根据工作流ID获取工作流的详细信息
// @Tags 工作流管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} response.Response{data=ent.Workflow} "工作流详情"
// @Failure 400 {object} response.Response "无效的工作流ID"
// @Failure 500 {object} response.Response "获取工作流失败"
// @Router /workflow/workflows/{id} [get]
func (c *Controller) GetWorkflowByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的工作流ID")
		return
	}
	uid := ctx.GetInt("userID")

	workflow, err := c.workflowMgr.GetWorkflowByID(ctx, id, uid)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取工作流失败: "+err.Error())
		return
	}

	response.Success(ctx, workflow)
}

// ListWorkflows 获取工作流列表
// @Summary 获取工作流列表
// @Description 获取当前用户的工作流列表，支持分页和筛选
// @Tags 工作流管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param page query int false "页码，默认1"
// @Param page_size query int false "每页条数，默认10"
// @Param start_date query string false "开始日期，格式：2006-01-02"
// @Param end_date query string false "结束日期，格式：2006-01-02"
// @Param status query string false "状态筛选，可选值：pending|running|completed|failed|cancelled"
// @Param domain query string false "域名筛选，支持模糊匹配"
// @Success 200 {object} response.Response{data=[]ent.Workflow} "工作流列表"
// @Failure 500 {object} response.Response "获取工作流列表失败"
// @Router /workflow/workflows [get]
func (c *Controller) ListWorkflows(ctx *gin.Context) {
	userID := ctx.GetInt("userID")
	params := pagination.ParsePaginationParams(
		ctx.DefaultQuery("page", "1"),
		ctx.DefaultQuery("pageSize", "10"),
	)

	// 获取筛选参数
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	status := ctx.Query("status")
	domain := ctx.Query("domain")

	// 如果传入了域名参数，先尝试解析为 punycode（国际化域名）
	if domain != "" {
		domain = strings.ToLower(strings.TrimSpace(domain))
	}
	workflows, total, err := c.workflowMgr.ListWorkflows(ctx, userID, params.Page, params.PageSize, startDate, endDate, status, domain)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "获取工作流列表失败: "+err.Error())
		return
	}

	response.Success(ctx, pagination.NewPaginatedResponse(
		workflows,
		total,
		params.Page,
		params.PageSize,
	))
}

// UpdateWorkflowStatus 更新工作流状态
// @Summary 更新工作流状态
// @Description 更新指定工作流的状态
// @Tags 工作流管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Param request body map[string]interface{} true "状态更新请求"
// @Success 200 {object} response.Response "工作流状态已更新"
// @Failure 400 {object} response.Response "无效的工作流ID或请求参数"
// @Failure 500 {object} response.Response "更新工作流状态失败"
// @Router /workflow/workflows/{id}/status [put]
func (c *Controller) UpdateWorkflowStatus(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的工作流ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的请求参数")
		return
	}

	if err := c.workflowMgr.UpdateWorkflowStatus(ctx, id, req.Status); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "更新工作流状态失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{"message": "工作流状态已更新"})
}

// StartWorkflow 启动工作流
// @Summary 启动工作流
// @Description 根据工作流ID启动工作流执行
// @Tags 工作流管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} response.Response "工作流已成功启动"
// @Failure 400 {object} response.Response "无效的工作流ID"
// @Failure 500 {object} response.Response "启动工作流失败"
// @Router /workflow/workflows/{id}/start [post]
func (c *Controller) StartWorkflow(ctx *gin.Context) {
	workflowID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的工作流ID"})
		return
	}
	uid := ctx.GetInt("userID")
	err = c.workflowMgr.StartWorkflow(ctx, workflowID, uid)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "启动工作流失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{"message": "工作流已成功启动"})
}

// RetryWorkflow 重试工作流
// @Summary 重试工作流
// @Description 重试失败的工作流
// @Tags 工作流管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} response.Response "工作流已重新启动"
// @Failure 400 {object} response.Response "无效的工作流ID"
// @Failure 500 {object} response.Response "重试工作流失败"
// @Router /workflow/workflows/{id}/retry [post]
func (c *Controller) RetryWorkflow(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的工作流ID")
		return
	}
	uid := ctx.GetInt("userID")

	if err := c.workflowMgr.RetryWorkflow(ctx, id, uid); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "重试工作流失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{"message": "工作流已重新启动"})
}

// CancelWorkflow 取消工作流
// @Summary 取消工作流
// @Description 取消正在运行的工作流
// @Tags 工作流管理
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "工作流ID"
// @Success 200 {object} response.Response "工作流已取消"
// @Failure 400 {object} response.Response "无效的工作流ID"
// @Failure 500 {object} response.Response "取消工作流失败"
// @Router /workflow/workflows/{id}/cancel [post]
func (c *Controller) CancelWorkflow(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的工作流ID")
		return
	}
	uid := ctx.GetInt("userID")
	if err := c.workflowMgr.CancelWorkflow(ctx, id, uid); err != nil {
		response.Error(ctx, http.StatusInternalServerError, "取消工作流失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{"message": "工作流已取消"})
}
