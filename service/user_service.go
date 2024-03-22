package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
)

// UserRegister
// @Summary 用户注册
// @Accept json
// @Produce json
// @Param userRegisterReq body api.UserRegisterReq true "注册请求参数"
// @Success 200 {object} response.RespData "无响应数据"
// @Router /user/register [POST]
func UserRegister(c *gin.Context) {
	writer := c.Writer
	var req api.UserRegisterReq
	err := c.ShouldBind(&req) // Form表单
	if err != nil {
		response.RespBadReq(writer, "参数不正确")
		return
	}
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 查询是否已注册
		_, isExist := models.FindUserByPhone(tx, req.Telephone)
		if isExist == true {
			return errors.New("手机号已注册")
		}
		// 用户不存在，创建用户记录
		salt := common.MakeSalt()
		userId := common.GenerateUUID()
		if err := tx.Create(&models.UserBasic{
			Username:         req.Username,
			Password:         common.MakePassword(req.Password, salt),
			Salt:             salt,
			Phone:            req.Telephone,
			UserId:           userId,
			UserType:         1,
			TotalStorageSize: 10240000000, // 字节为单位
			StorageSize:      0,
		}).Error; err != nil {
			return err
		}
		// 创建用户网盘根目录记录
		if err := tx.Create(&models.UserRepository{
			UserFileId: common.GenerateUUID(),
			FileName:   "/",
			UserId:     userId,
			FileType:   filehandler.DIRECTORY,
			IsDir:      1,
		}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		response.RespOK(writer, response.DatabaseError, false, nil, err.Error())
		return
	}
	response.RespOK(writer, response.Success, true, nil, "注册成功")
	return
}

// UserLogin
// @Summary 用户登录，并返回cookie。
// @Accept json
// @Produce json
// @Param userLoginReq query api.UserLoginReq true "请求参数"
// @Success 200 {object} response.RespData{data=api.UserLoginResp} "cookie"
// @Router /user/login [GET]
func UserLogin(c *gin.Context) {
	writer := c.Writer
	// 解析请求的query参数
	var req api.UserLoginReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespBadReq(writer, "参数错误")
		return
	}
	// 查询用户是否存在
	ub, isExist := models.FindUserByPhone(models.DB, req.Telephone)
	if !isExist {
		response.RespOK(writer, response.UserNotExist, false, nil, "用户不存在")
		return
	}
	// 校验密码
	isPass := common.ValidatePassword(req.Password, ub.Salt, ub.Password)
	if !isPass {
		response.RespOK(writer, response.WrongPassword, false, nil, "密码错误")
		return
	}
	// 生成token
	token, err := common.GenerateCookie(ub.Username, req.Telephone, ub.UserId)
	if err != nil {
		response.RespOK(writer, response.CookieGenError, false, nil, "cookie签发失败")
		return
	}
	response.RespOK(writer, 0, true, api.UserLoginResp{Token: token}, "登陆成功")
}

// CheckLogin
// @Summary 检查用户是否登录，并返回用户名，用户id。
// @Accept json
// @Produce json
// @Success 200 {object} response.RespData{data=api.UserCheckLoginResp} "响应"
// @Router /user/checkuserlogininfo [GET]
func CheckLogin(c *gin.Context) {
	writer := c.Writer
	cookie, err := c.Cookie("token")
	if err != nil {
		response.RespOK(writer, response.CookieNotValid, false, nil, "cookie校验失败")
		return
	}
	// 解析出userClaim
	userClaim, err := common.ParseCookie(cookie)
	if err != nil {
		response.RespOK(writer, response.CookieNotValid, false, nil, "cookie校验失败")
		return
	}
	response.RespOK(writer, 0, true, api.UserCheckLoginResp{
		UserId:   userClaim.UserId,
		UserName: userClaim.Username,
	}, "cookie校验成功")
}
