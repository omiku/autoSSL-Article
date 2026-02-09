package auth

//	@securityDefinitions.http
//	@name			ApiKeyAuth
//	@type			http
//	@scheme			bearer
//	@bearerFormat	JWT

import (
	"log"
	"net/http"
	"time"

	"autoSSL/api/response"
	"autoSSL/service/rbac"
	"autoSSL/utils/jwt"

	"github.com/gin-gonic/gin"
)

type Claims struct {
	UserID uint `json:"user_id"`
}

// JWTMiddleware 接收JWTManager实例的中间件构造函数
func JWTMiddleware(jwtUtil *jwt.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未提供访问令牌"})
			return
		}

		// 使用JWTManager的ValidateToken方法验证令牌
		claims, err := jwtUtil.ValidateToken(tokenString)
		if err != nil || claims.TokenType != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "无效令牌: " + err.Error()})
			return
		}

		// 检查令牌是否临近过期需要刷新
		if time.Until(claims.ExpiresAt.Time) < 30*time.Minute {
			c.Header("X-Token-Expired", "true")
		}
		c.Set("userID", claims.UserID)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

// RoleMiddleware 创建一个基于RBAC的权限检查中间件
// 统一使用用户ID进行权限验证，通过Casbin策略中的用户-角色关联来管理权限
func RoleMiddleware(rbacService *rbac.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 从JWT上下文中获取用户ID
		userIDValue, exists := ctx.Get("userID")
		if !exists {
			log.Printf("未找到用户ID信息")
			response.Error(ctx, http.StatusForbidden, "权限不足：未找到用户ID")
			ctx.Abort()
			return
		}

		userID, ok := userIDValue.(int)
		if !ok || userID <= 0 {
			log.Printf("用户ID信息无效: %v", userIDValue)
			response.Error(ctx, http.StatusForbidden, "权限不足：用户ID无效")
			ctx.Abort()
			return
		}

		// 获取请求的路径和方法
		requestPath := ctx.Request.URL.Path
		requestMethod := ctx.Request.Method

		hasPermission, err := rbacService.Enforce(ctx.Request.Context(), userID, requestPath, requestMethod)
		if err != nil {
			log.Printf("权限检查失败: %v", err)
			response.Error(ctx, http.StatusInternalServerError, err.Error())
			ctx.Abort()
			return
		}

		if !hasPermission {
			response.Error(ctx, http.StatusForbidden, "权限不足")
			ctx.Abort()
			return
		}

		// 权限检查通过，继续处理请求
		ctx.Next()
	}
}
