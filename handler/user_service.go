package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	ApiModels "netdisk_in_go/APImodels"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
)

// UserRegister 用户注册
func UserRegister(c *gin.Context) {
	writer := c.Writer
	var req ApiModels.UserRegisterApi
	err := c.ShouldBind(&req)
	if err != nil {
		utils.RespBadReq(writer, "参数不正确")
		return
	}
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		// 用户注册时，查看用户注册电话是否存在
		_, isExist := models.FindUserByPhone(tx, req.Telephone)
		if isExist == true {
			return errors.New("手机号已注册")
		}
		// 用户不存在，开始注册
		salt := utils.MakeSalt()
		ub := models.UserBasic{
			Username:         req.Username,
			Password:         utils.MakePassword(req.Password, salt),
			Salt:             salt,
			Phone:            req.Telephone,
			UserId:           utils.GenerateUUID(),
			UserType:         1,
			TotalStorageSize: 10240000000,
			StorageSize:      0,
		}
		if err := tx.Create(&ub).Error; err != nil {
			return errors.New("注册失败，请联系管理员")
		}
		ur := models.UserRepository{
			UserFileId: utils.GenerateUUID(),
			FileName:   "/",
			UserId:     ub.UserId,
			FileType:   utils.DIRECTORY,
			IsDir:      1,
		}
		if err := tx.Create(&ur).Error; err != nil {
			return errors.New("注册失败，请联系管理员")
		}
		utils.RespOK(writer, 0, true, nil, "注册成功")
		return nil
	})
	if err != nil {
		utils.RespBadReq(writer, err.Error())
	}
}

// UserLogin 用户登录
func UserLogin(c *gin.Context) {
	writer := c.Writer
	phone := c.Query("telephone")
	rawPassword := c.Query("password")
	// 查询用户是否存在
	ub, isExist := models.FindUserByPhone(utils.DB, phone)
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
	token, err := utils.GenerateToken(ub.Username, phone, ub.UserId, 360000)
	if err != nil {
		utils.RespBadReq(writer, "登陆失败，请联系管理员")
	}
	utils.RespOK(writer, 0, true, gin.H{"token": token}, "登陆成功")
}

// CheckLogin 检查用户是否登录
func CheckLogin(c *gin.Context) {
	writer := c.Writer
	uc, err := utils.ParseCookieFromRequest(c)
	if err != nil {
		utils.RespOK(writer, 999999, false, nil, "未登录")
		return
	}
	utils.RespOK(writer, 0, true, gin.H{
		"userId":   uc.UserId,
		"username": uc.Username,
	}, "成功") // todo:用户信息存于data中
}
