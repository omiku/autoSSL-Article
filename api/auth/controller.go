package auth

import (
	"autoSSL/api/response"
	"autoSSL/logger"
	"autoSSL/service/auth"
	"autoSSL/utils/jwt"
	"autoSSL/utils/security"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthController struct {
	userService *auth.Service
	jwtUtil     *jwt.JWTManager
}

func NewAuthController(userService *auth.Service, jwtUtil *jwt.JWTManager) *AuthController {
	return &AuthController{
		userService: userService,
		jwtUtil:     jwtUtil,
	}
}

// LoginRequest 登录请求参数结构体
// @Description 登录接口所需的请求参数
type LoginRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	CaptchaID string `json:"captcha_id" binding:"required"`
	Answer    string `json:"answer" binding:"required"`
}

// LoginResponse 登录响应结构体
// @Description 登录接口返回的响应数据
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// @Summary		用户登录
// @Description	通过邮箱、密码和验证码登录系统获取访问令牌
// @Tags			认证
// @Accept			json
// @Produce		json
// @Param			login	body		auth.LoginRequest	true	"登录请求参数"
// @Success		200		{object}	auth.LoginResponse	"登录成功响应（包含access_token和refresh_token）"
// @Failure		400		{string}	string				"无效请求格式"
// @Failure		401		{string}	string				"验证码错误/用户不存在/密码错误"
// @Failure		500		{string}	string				"服务器错误"
// @Router			/auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效请求格式")
		return
	}

	if !c.userService.VerifyCaptcha(req.CaptchaID, req.Answer) {
		response.Error(ctx, http.StatusBadRequest, "验证码错误")
		return
	}

	user, err := c.userService.GetUserByEmail(ctx, req.Email)
	if err != nil || !security.CheckPassword(req.Password, user.PasswordHash) {
		response.Error(ctx, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	// 获取用户角色用于生成token
	roles, err := c.userService.GetUserRoles(ctx, user.ID)
	if err != nil {
		logger.Error("获取用户角色失败:", zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "获取用户权限失败")
		return
	}

	// 将角色转换为字符串数组
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	accessToken, refreshToken, err := c.GenerateTokens(user.ID, roleNames)
	if err != nil {
		logger.Error("Token generation failed:", zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "令牌生成失败")
		return
	}

	response.Success(ctx, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// RegisterRequest 注册请求参数结构体
//
//	@Description	注册接口所需的请求参数
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	CaptchaID string `json:"captcha_id" binding:"required"`
	Answer    string `json:"answer" binding:"required"`
}

// RegisterResponse 注册响应结构体
//
//	@Description	注册接口返回的响应数据
type RegisterResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

// RefreshRequest 刷新令牌请求参数结构体
// @Description 刷新令牌接口所需的请求参数
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshResponse 刷新令牌响应结构体
// @Description 刷新令牌接口返回的响应数据
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌和刷新令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param refresh body RefreshRequest true "刷新令牌请求参数"
// @Success 200 {object} RefreshResponse "刷新成功响应（包含新的access_token和refresh_token）"
// @Failure 400 {string} string "无效请求格式"
// @Failure 401 {string} string "无效刷新令牌"
// @Failure 500 {string} string "服务器错误"
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	var req RefreshRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request format:", zap.Error(err))
		response.Error(ctx, http.StatusBadRequest, "无效请求格式")
		return
	}

	claims, err := c.parseToken(req.RefreshToken)
	if err != nil {
		logger.Error("Invalid refresh token:", zap.Error(err))
		response.Error(ctx, http.StatusUnauthorized, "无效刷新令牌")
		return
	}
	if claims.TokenType == "access" {
		response.Error(ctx, http.StatusBadRequest, "无效的刷新令牌类型")
		return
	}

	// 获取用户角色用于生成新令牌
	roles, err := c.userService.GetUserRoles(ctx, claims.UserID)
	if err != nil {
		logger.Error("获取用户角色失败:", zap.Error(err))
		response.Error(ctx, http.StatusInternalServerError, "获取用户权限失败")
		return
	}

	// 将角色转换为字符串数组
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	newAccessToken, newRefreshToken, err := c.GenerateTokens(claims.UserID, roleNames)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "令牌生成失败")
		return
	}
	response.Success(ctx, RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}

func (c *AuthController) GenerateTokens(userID int, roles []string) (string, string, error) {
	// 生成新的访问令牌和刷新令牌
	accessToken, err := c.jwtUtil.GenerateAccessToken(userID, roles)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := c.jwtUtil.GenerateRefreshToken(userID, roles)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (c *AuthController) parseToken(tokenString string) (*jwt.Claims, error) {
	return c.jwtUtil.ValidateToken(tokenString)
}

// Register 用户注册
// @Summary 用户注册
// @Description 通过邮箱、密码和验证码注册新用户（第一个用户自动为管理员）
// @Tags 认证
// @Accept json
// @Produce json
// @Param register body RegisterRequest true "注册请求参数"
// @Success 201 {object} RegisterResponse "注册成功响应（包含用户ID和邮箱）"
// @Failure 400 {string} string "无效请求格式/验证码错误"
// @Failure 409 {string} string "用户已存在"
// @Failure 500 {string} string "服务器错误"
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid request format:", zap.Error(err))
		response.Error(ctx, http.StatusBadRequest, "无效请求格式")
		return
	}

	if !c.userService.VerifyCaptcha(req.CaptchaID, req.Answer) {
		response.Error(ctx, http.StatusBadRequest, "验证码错误")
		return
	}

	user, err := c.userService.CreateUser(ctx, req.Email, req.Password, "user")
	if err != nil {
		response.Error(ctx, http.StatusConflict, "用户已存在")
		return
	}

	response.Success(ctx, RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}

type CaptchaResponse struct {
	CaptchaID string `json:"captcha_id"`
	Image     any    `json:"image"`
	Error     error  `json:"error"`
}

// GenerateCaptcha 生成验证码
// @Summary 生成验证码
// @Description 生成图片验证码用于登录/注册验证
// @Tags 认证
// @Produce json
// @Success 200 {object} CaptchaResponse "验证码响应（包含captcha_id和image）"
// @Failure 500 {string} string "服务器错误"
// @Router /auth/captcha [get]
func (c *AuthController) GenerateCaptcha(ctx *gin.Context) {
	id, b64s, err := c.userService.GenerateCaptcha()
	captchaResponse := CaptchaResponse{
		CaptchaID: id,
		Image:     b64s,
		Error:     err,
	}
	response.Success(ctx, captchaResponse)
}
