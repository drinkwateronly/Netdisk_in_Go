package service

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
)

// UserRegister 用户注册
func UserRegister(c *gin.Context) {
	json := make(map[string]interface{}) //注意该结构接受的内容
	c.BindJSON(&json)

	phone := json["telephone"].(string)
	// 用户注册时，查看用户注册电话是否存在
	_, isExist := models.FindUserByPhone(phone)
	if isExist {
		c.JSON(400, gin.H{
			"message": "手机号已注册",
		})
		return
	}

	ub := &models.UserBasic{}
	// 用户不存在，开始注册
	ub.Username = json["username"].(string)
	ub.Phone = phone
	// 密码加盐
	salt := utils.MakeSalt()
	ub.Salt = salt
	rawPassword := json["password"].(string)

	ub.Password = utils.MakePassword(rawPassword, salt)

	res := models.CreateUser(ub)
	if res.Error != nil {
		c.JSON(400, gin.H{
			"message": "注册失败",
			"success": false,
		})
		return
	}
	c.JSON(200, gin.H{
		"success": false,
		"message": "注册成功",
	})
}

// UserLogin 用户登录
func UserLogin(c *gin.Context) {
	phone := c.Query("telephone")
	rawPassword := c.Query("password")
	// 查询用户是否存在
	ub, isExist := models.FindUserByPhone(phone)
	if !isExist {
		c.JSON(400, gin.H{
			"success": false,
			"message": "用户不存在",
		})
		return
	}
	// 校验密码
	isPass := utils.ValidatePassword(rawPassword, ub.Salt, ub.Password)
	if !isPass {
		c.JSON(400, gin.H{
			"success": false,
			"message": "密码错误",
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "登陆成功",
	})
}
