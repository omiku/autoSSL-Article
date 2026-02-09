package auth

import (
	"autoSSL/bootstrap"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.RouterGroup, container *bootstrap.Container) {
	// 初始化 AuthService
	authGroup := NewAuthController(container.AuthService, container.JWTUtil)

	r.POST("/login", authGroup.Login)
	r.POST("/register", authGroup.Register)
	r.POST("/refresh", authGroup.RefreshToken)
	r.GET("/captcha", authGroup.GenerateCaptcha)
}
