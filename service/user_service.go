package service

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
)

// UserRegister 用户注册
func UserRegister(c *gin.Context) {
	writer := c.Writer

	json := make(map[string]interface{}) //注意该结构接受的内容
	c.BindJSON(&json)
	phone := json["telephone"].(string)
	// 用户注册时，查看用户注册电话是否存在
	_, isExist := models.FindUserByPhone(phone)

	if isExist == true {
		utils.RespBadReq(writer, "手机号已注册")
		return
	}

	ub := &models.UserBasic{}
	// 用户不存在，开始注册
	ub.Username = json["username"].(string)
	ub.Phone = phone
	// 密码加盐
	salt := utils.MakeSalt()
	ub.Salt = salt
	ub.UserId = utils.GenerateUUID()
	ub.UserType = 1
	ub.TotalStorageSize = 1024000000
	rawPassword := json["password"].(string)

	ub.Password = utils.MakePassword(rawPassword, salt)

	res := models.CreateUser(ub)
	if res.Error != nil {
		utils.RespBadReq(writer, "注册失败，请联系管理员")
		return
	}
	utils.RespOK(writer, 0, true, nil, "注册成功")
}

// UserLogin 用户登录
func UserLogin(c *gin.Context) {
	writer := c.Writer
	phone := c.Query("telephone")
	rawPassword := c.Query("password")
	// 查询用户是否存在
	ub, isExist := models.FindUserByPhone(phone)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
		return
	}
	// 校验密码
	isPass := utils.ValidatePassword(rawPassword, ub.Salt, ub.Password)
	if !isPass {
		utils.RespBadReq(writer, "密码错误")
		return
	}
	// 生成token
	token, err := utils.GenerateToken(phone, ub.UserId, 360000)
	if err != nil {
		utils.RespBadReq(writer, "登陆失败，请联系管理员")
	}
	utils.RespOK(writer, 0, true, struct {
		Token string `json:"token"`
	}{Token: token}, "登陆成功")
}

// CheckLogin 检查用户是否登录
func CheckLogin(c *gin.Context) {
	writer := c.Writer
	_, err := utils.ParseCookieFromRequest(c)
	if err != nil {
		utils.RespOK(writer, 999999, false, nil, "未登录")
		return
	}
	utils.RespOK(writer, 0, true, nil, "成功") // todo:用户信息存于data中
}
