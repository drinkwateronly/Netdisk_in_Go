package middleware

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/common"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
)

// Authentication 身份认证中间件
func Authentication(c *gin.Context) {
	writer := c.Writer
	cookie, err := c.Cookie("token")
	if err != nil {
		response.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}
	// 解析出userClaim
	userClaim, err := common.ParseCookie(cookie)
	if err != nil {
		response.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}
	// 根据userClaim的用户id查询用户基本信息
	var userBasic models.UserBasic
	res := models.DB.Where("user_id = ?", userClaim.UserId).Find(&userBasic)
	if res.Error != nil || res.RowsAffected == 0 {
		response.RespOK(writer, 999999, false, nil, "用户不存在")
		return
	}
	c.Set("userBasic", &userBasic)
	c.Next()
}
