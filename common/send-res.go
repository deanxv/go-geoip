package common

import (
	"github.com/gin-gonic/gin"
)

func SendResponse(c *gin.Context, httpCode int, code int, message string, data interface{}) {
	c.JSON(httpCode, NewResponseResult(code, message, data))
}
