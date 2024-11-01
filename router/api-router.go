package router

import (
	"github.com/gin-gonic/gin"
	"ip2region-geoip/controller"
	"ip2region-geoip/middleware"
)

func SetApiRouter(router *gin.Engine) {

	// 全局 Middlewares
	router.Use(middleware.CORS())
	router.Use(middleware.RequestRateLimit())

	// 启用身份验证中间件
	router.Use(middleware.Auth())

	// 无需身份验证的路由
	router.GET("/ip", controller.IpNoArgs)
	router.GET("/ip/:ip", controller.Ip)
}
