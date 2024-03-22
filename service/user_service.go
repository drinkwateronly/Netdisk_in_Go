package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/models"
	ApiModels "netdisk_in_go/models/api_models"
	"netdisk_in_go/utils"
)

// https://www.jb51.net/article/259993.htm

// UserRegister
// @Summary 用户注册
// @Accept json
// @Produce json
// @Param telephone body string true "用户电话"
// @Param username body string true "用户名"
// @Param password body string true "密码"
// @Success 200 {object} api_models.RespData{} ""
// @Failure 400 {object} string "参数出错"
// @Router /user/register [POST]
func UserRegister(c *gin.Context) {
	writer := c.Writer
	var req ApiModels.UserRegisterReqAPI
	err := c.ShouldBind(&req) // Form表单
	if err != nil {
		utils.RespBadReq(writer, "参数不正确")
		return
	}
	err = models.DB.Transaction(func(tx *gorm.DB) error {
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
			return err
		}
		ur := models.UserRepository{
			UserFileId: utils.GenerateUUID(),
			FileName:   "/",
			UserId:     ub.UserId,
			FileType:   utils.DIRECTORY,
			IsDir:      1,
		}
		if err := tx.Create(&ur).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.RespOK(writer, 99999, false, nil, err.Error())
		return
	}
	utils.RespOK(writer, 0, true, nil, "注册成功")
	return
}

// UserLogin
// @Summary 用户登录，并返回cookie。
// @Accept json
// @Produce json
// @Param telephone query string true "用户电话"
// @Param password query string true "密码"
// @Success 200 {object} api_models.RespData{data=api_models.UserLoginRespAPI} "cookie"
// @Failure 400 {object} string "参数出错"
// @Router /user/login [GET]
func UserLogin(c *gin.Context) {
	writer := c.Writer
	// 解析请求的query参数
	var req ApiModels.UserLoginReqAPI
	err := c.ShouldBindQuery(&req)
	if err != nil {
		utils.RespBadReq(writer, "参数错误")
		return
	}

	// 查询用户是否存在
	ub, isExist := models.FindUserByPhone(models.DB, req.Telephone)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
		return
	}
	// 校验密码
	isPass := utils.ValidatePassword(req.Password, ub.Salt, ub.Password)
	if !isPass {
		utils.RespBadReq(writer, "密码错误")
		return
	}
	// 生成token
	token, err := utils.GenerateToken(ub.Username, req.Telephone, ub.UserId)
	if err != nil {
		utils.RespBadReq(writer, "登陆失败，请联系管理员")
		return
	}
	utils.RespOK(writer, 0, true, ApiModels.UserLoginRespAPI{Token: token}, "登陆成功")
}

// CheckLogin
// @Summary 检查用户是否登录，并返回用户名，用户id。
// @Accept json
// @Produce json
// @Success 200 {object} api_models.RespData{data=api_models.UserCheckLoginRespAPI} "cookie"
// @Failure 400 {object} string "参数出错"
// @Router /user/checkuserlogininfo [GET]
func CheckLogin(c *gin.Context) {
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
	utils.RespOK(writer, 0, true, ApiModels.UserCheckLoginRespAPI{
		UserId:   userClaim.UserId,
		UserName: userClaim.Username,
	}, "成功")
}
