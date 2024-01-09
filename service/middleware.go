package service

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/utils"
)

// Authentication 身份认证中间件
func Authentication(c *gin.Context) {
	writer := c.Writer
	_, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}
	c.Next()
}
