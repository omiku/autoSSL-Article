package main

//	@title			AutoSSL API
//	@version		1.0
//	@description		AutoSSL 证书管理API文档
//	@termsOfService			http://example.com/terms
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in			header
//	@name		Authorization

//	@contact.name	API Support
//	@contact.email	support@example.com

//	@license.name	MIT
//	@license.url	http://opensource.org/licenses/MIT

//	@host			localhost:8080
//	@basePath		/apiv1
//	@query.collection.format	multi

import (
	"autoSSL/api"
	"autoSSL/bootstrap"
	"autoSSL/logger"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 通过bootstrap初始化所有依赖
	container, err := bootstrap.Init("./config.yaml")
	if err != nil {
		log.Fatalf("应用初始化失败: %v", err)
	}
	defer container.Close()

	// 创建Gin实例当日志不为debug时设置为release模式
	if container.Config.Logger.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 设置路由并注入Container依赖
	api.SetupRoutes(r, container)

	// 启动服务
	if err := r.Run(fmt.Sprintf(":%d", container.Config.Server.Port)); err != nil {
		logger.Fatal("服务启动失败", zap.Error(err))
	}
}
