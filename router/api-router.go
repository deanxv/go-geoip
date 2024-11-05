package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go-geoip/common/config"
	"go-geoip/controller"
	_ "go-geoip/docs"
	"go-geoip/middleware"
)

func SetApiRouter(router *gin.Engine) {

	// 全局 Middlewares
	router.Use(middleware.CORS())
	router.Use(middleware.RequestRateLimit())

	if config.SwaggerEnable == "" || config.SwaggerEnable == "1" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 启用身份验证中间件
	router.Use(middleware.Auth())

	// 无需身份验证的路由
	router.GET("/ip", controller.IpNoArgs)
	router.GET("/ip/:ip", controller.Ip)
}
