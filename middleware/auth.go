package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"ip2region-geoip/common"
	"ip2region-geoip/common/config"
	"net/http"
	"strings"
)

func isValidSecret(secret string) bool {
	return config.ApiSecret != "" && !lo.Contains(config.ApiSecrets, secret)
}

func authHelper(c *gin.Context) {
	secret := c.Request.Header.Get("Authorization")
	secret = strings.Replace(secret, "Bearer ", "", 1)
	if isValidSecret(secret) {
		common.SendResponse(c, http.StatusUnauthorized, 1, "auth fail", nil)
		c.Abort()
		return
	}

	if config.ApiSecret == "" {
		c.Request.Header.Set("Authorization", "")
	}

	c.Next()
	return
}

func Auth() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHelper(c)
	}
}
