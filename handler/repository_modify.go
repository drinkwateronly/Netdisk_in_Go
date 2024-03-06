package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
	"strings"
)

// MoveFile 文件
func MoveFile(c *gin.Context) {
	type MoveFileRequest struct {
		FilePath   string `json:"filePath"`
		UserFileId string `json:"userFileId"`
	}
	ub, err := models.GetUserFromCoookie(utils.DB, c)
	writer := c.Writer
	if err != nil {
		utils.RespBadReq(writer, "用户不存在")
		return
	}
	var req MoveFileRequest
	err = c.ShouldBind(&req)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}

	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		ur, isExist := models.FindUserFileById(ub.UserId, req.UserFileId)
		if !isExist {
			return errors.New("文件不存在")
		}
		// 找要移动的目录有无同
		res := utils.DB.Where("user_id = ? AND file_path = ? AND file_name = ? AND is_dir = 0", ub.UserId, req.FilePath, ur.FileName)
		if res.RowsAffected != 0 {
			return errors.New("该目录下同名文件已存在")
		}
		if err != nil {
			return err
		}
		if err := utils.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, req.UserFileId).
			Updates(&models.UserRepository{FilePath: req.FilePath}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.RespOK(writer, 999999, false, nil, err.Error())
		return
	}
	utils.RespOK(writer, 0, true, nil, "文件移动成功")
}

// RenameFile 文件重命名
func RenameFile(c *gin.Context) {
	type RenameFileRequest struct {
		FileName   string `json:"fileName"`
		UserFileId string `json:"userFileId"`
	}
	ub, err := models.GetUserFromCoookie(utils.DB, c)
	writer := c.Writer
	if err != nil {
		utils.RespBadReq(writer, "用户不存在")
		return
	}
	var req RenameFileRequest
	err = c.ShouldBind(&req)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	if strings.ContainsAny(req.FileName, "|<>/\\:*?\"") {
		utils.RespOK(writer, 999999, false, nil, "命名失败，文件名称出现非法字符")
		return
	}
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		ur, isExist := models.FindUserFileById(ub.UserId, req.UserFileId)
		if !isExist {
			return errors.New("文件不存在")
		}
		res := utils.DB.Where("user_id = ? AND file_path = ? AND file_name = ? AND user_file_id != ?", ub.UserId, ur.FilePath, req.FileName, req.UserFileId)
		if res.RowsAffected != 0 {
			return errors.New("该目录下同名文件已存在")
		}
		if err != nil {
			return err
		}
		if err := utils.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, req.UserFileId).Updates(models.UserRepository{FileName: req.FileName}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.RespOK(writer, 999999, false, nil, err.Error())
		return
	}
	utils.RespOK(writer, 0, true, nil, "文件名修改成功")
}
