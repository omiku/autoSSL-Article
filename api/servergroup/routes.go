package servergroup

import (
	"autoSSL/bootstrap"
	"autoSSL/service/servergroup"

	"github.com/gin-gonic/gin"
)

func getServerGroupService() *servergroup.ServerGroupService {
	return &servergroup.ServerGroupService{}
}

func SetupRoutes(r *gin.RouterGroup, container *bootstrap.Container) {
	service := container.ServerGroupSvc
	controller := NewServerGroupController(service)

	r.POST("/servers", controller.Create)
	r.GET("/servers", controller.Get)
	r.PUT("/servers/:id", controller.Update)
	r.DELETE("/servers/:id", controller.Delete)
}
