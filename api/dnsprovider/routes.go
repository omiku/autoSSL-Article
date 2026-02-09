package dnsprovider

import (
	"autoSSL/bootstrap"
	dnspd "autoSSL/service/dnsprovider"

	"github.com/gin-gonic/gin"
)

func getDNService() *dnspd.Service {
	// 实际实现中需要正确初始化 DNSProviderService
	return &dnspd.Service{}
}

func SetupRoutes(r *gin.RouterGroup, container *bootstrap.Container) {

	dnsprovider := container.DNSService
	dnsGroup := NewDNSProviderController(dnsprovider)

	r.POST("/dnsproviders", dnsGroup.Create)
	r.GET("/dnsproviders", dnsGroup.GetAll)
	r.GET("/dnsproviders/:id", dnsGroup.Get)
	r.PUT("/dnsproviders/:id", dnsGroup.Update)
	r.DELETE("/dnsproviders/:id", dnsGroup.Delete)
}
