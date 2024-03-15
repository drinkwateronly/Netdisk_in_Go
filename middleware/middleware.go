package middleware

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
)

// Authentication 身份认证中间件
func Authentication(c *gin.Context) {
	writer := c.Writer
	cookie, err := c.Cookie("token")
	if err != nil {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}
	// 解析出userClaim
	userClaim, err := utils.ParseCookie(cookie)
	if err != nil {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}
	// 根据userClaim的用户id查询用户基本信息
	var userBasic models.UserBasic
	res := utils.DB.Where("user_id = ?", userClaim.UserId).Find(&userBasic)
	if res.Error != nil || res.RowsAffected == 0 {
		utils.RespOK(writer, 999999, false, nil, "用户不存在")
		return
	}
	c.Set("userBasic", &userBasic)
	c.Next()
}
