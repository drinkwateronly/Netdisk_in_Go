package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
	"strings"
)

//  回收站文件相关

// GetRecoveryFileList 1
func GetRecoveryFileList(c *gin.Context) {
	writer := c.Writer
	// 校验cookie，获取用户信息
	ub, err := models.GetUserFromCoookie(utils.DB, c)
	if err != nil {
		utils.RespBadReq(writer, "用户校验失败")
		return
	}
	var recoveryFiles []models.RecoveryBasic
	utils.DB.
		Where("user_id = ?", ub.UserId).
		Find(&recoveryFiles)
	utils.RespOkWithDataList(writer, 0, recoveryFiles, len(recoveryFiles), "文件列表")
}

func DelRecoveryFile(c *gin.Context) {
	writer := c.Writer
	// 校验cookie，获取用户信息
	ub, err := models.GetUserFromCoookie(utils.DB, c)
	if err != nil {
		utils.RespBadReq(writer, "用户校验失败")
		return
	}
	type DelRecoveryFileReq struct {
		UserFileId string `json:"userFileId"`
	}
	var r DelRecoveryFileReq
	err = c.ShouldBind(&r)
	if err != nil {
		utils.RespBadReq(writer, "参数错误")
		return
	}
	if err := utils.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, r.UserFileId).
		Delete(&models.RecoveryBasic{}).Error; err != nil {
		utils.RespOK(writer, 0, false, nil, "文件删除失败")
		return
	}
	utils.RespOkWithDataList(writer, 0, nil, 0, "删除成功")
}

func DelRecoveryFileInBatch(c *gin.Context) {
	writer := c.Writer
	// 校验cookie，获取用户信息
	ub, err := models.GetUserFromCoookie(utils.DB, c)
	if err != nil {
		utils.RespBadReq(writer, "用户校验失败")
		return
	}
	type DelRecoveryFileReq struct {
		UserFileIds string `json:"userFileIds"`
	}
	var r DelRecoveryFileReq
	err = c.ShouldBind(&r)
	if err != nil {
		utils.RespBadReq(writer, "参数错误")
		return
	}
	userFileIdList := strings.Split(r.UserFileIds, ",")
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		for _, userFileId := range userFileIdList {
			if err := utils.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, userFileId).
				Delete(&models.RecoveryBasic{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.RespOK(writer, 0, false, nil, "清空回收站失败")
		return
	}
	utils.RespOkWithDataList(writer, 0, nil, 0, "清空回收站成功")
}
