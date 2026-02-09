package certificate

import (
	"autoSSL/bootstrap"

	"github.com/gin-gonic/gin"
)

//func GetCertificateService() *certificate.CertificateService {
//	// 实际实现中需要正确初始化 CertificateService
//	return &certificate.CertificateService{}
//}

func SetupRoutes(r *gin.RouterGroup, container *bootstrap.Container) {
	getCertificateService := container.CertService
	certificateGroup := NewCertificateController(getCertificateService)

	r.POST("/certificates", certificateGroup.Create)
	r.GET("/certificates", certificateGroup.GetAll)
	r.GET("/certificates/:id", certificateGroup.Get)
	r.GET("/certificates/:id/download", certificateGroup.Download)
	//r.PUT("/certificates/:id", certificateGroup.Renew)
	//r.DELETE("/certificates/:id", deleteHandler)
}
