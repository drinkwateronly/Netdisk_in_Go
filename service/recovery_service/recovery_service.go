package recovery_service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"strings"
)

//  回收站文件相关接口

// GetRecoveryFileList
// @Summary 获取回收站文件列表
// @Accept json
// @Produce json
// @Param cookie query string true "Cookie" // 并非query参数
// @Success 200 {object} api_models.RespDataList{datalist=[]api_models.RecoveryListResp} "服务器响应成功，根据响应code判断是否成功"
// @Failure 400 {object} string "参数出错"
// @Router /recoveryfile/list [GET]
func GetRecoveryFileList(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	ub, boo := models.GetUserBasicFromContext(c)
	if !boo {
		response.RespUnAuthorized(writer)
	}
	var recoveryFiles []api.RecoveryListResp
	res := models.DB.Model(models.RecoveryBatch{}).Where("user_id = ?", ub.UserId).Scan(&recoveryFiles)
	if res.Error != nil {
		response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
		return
	}
	response.RespOkWithDataList(writer, response.Success, recoveryFiles, len(recoveryFiles), "回收站文件列表")
}

// DelRecoveryFile
// @Summary 删除单个回收站文件
// @Accept json
// @Produce json
// @Param userFileId body api.DelRecoveryReq true "用户文件id"
// @Success 200 {object} response.RespData "响应"
// @Router /recoveryfile/deleterecoveryfile [POST]
func DelRecoveryFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub, isExist := models.GetUserBasicFromContext(c)
	if !isExist {
		response.RespUnAuthorized(writer)
		return
	}
	// 绑定请求参数
	var r api.DelRecoveryReq
	err := c.ShouldBindJSON(&r)
	if err != nil {
		response.RespBadReq(writer, "参数错误")
		return
	}
	// 实际上可以直接通过recovery_batch表根据user_file_id删除，
	// 因为不会出现两条user_file_id一样且delete_at字段不为空的记录，即同一个文件不会同时被删除两次
	// 但由于user_file_id在该表中没有设置索引，因此改为查询两次表，两次都是带索引的查询。
	models.DB.Transaction(func(tx *gorm.DB) error {
		var ur models.UserRepository
		err := tx.Unscoped().Where("user_id = ? AND user_file_id = ?", ub.UserId, r.UserFileId).
			First(&ur).Error
		if err != nil { // 文件不存在
			response.RespOKFail(writer, response.RecoveryFileNotExist, "回收站文件不存在")
			return err
		}
		if len(ur.DeleteBatchId) == 0 {
			response.RespOKFail(writer, response.FileNotDeleted, "此文件未被删除")
			return err
		}
		err = models.DB.Where("delete_batch_id = ?", ur.DeleteBatchId).
			Delete(&models.RecoveryBatch{}).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}
		response.RespOKSuccess(writer, response.Success, nil, "删除成功")
		return nil
	})

}

// DelRecoveryInBatch
// @Summary 批量删除回收站文件
// @Accept json
// @Produce json
// @Param userFileId body api.DelRecoveryInBatchReq true "请求"
// @Success 200 {object} response.RespData "响应"
// @Router /recoveryfile/batchdelete [POST]
func DelRecoveryInBatch(c *gin.Context) {
	writer := c.Writer
	ub := c.MustGet("userBasic").(*models.UserBasic)
	// 解析请求参数
	var r api.DelRecoveryInBatchReq
	err := c.ShouldBindJSON(&r)
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
		return
	}
	userFileIdList := strings.Split(r.UserFileIds, ",")
	if len(userFileIdList) == 0 {
		response.RespBadReq(writer, "请求参数错误")
		return
	}

	models.DB.Transaction(func(tx *gorm.DB) error {
		var urs []models.UserRepository
		err := tx.Unscoped().Where("user_id = ? AND user_file_id IN ? AND delete_at <> 0", ub.UserId, userFileIdList).
			First(&urs).Error
		if err != nil { // 文件不存在
			response.RespOKFail(writer, response.FileNotExist, "有文件不存在或未被删除")
			return err
		}
		if len(urs) != len(userFileIdList) {
			response.RespOKFail(writer, response.FileNotExist, "有文件不存在或未被删除")
			return err
		}

		// 获取这些文件的delete_batch_id
		deleteBatchIds := make([]string, len(urs))
		for i := range urs {
			if len(urs[i].DeleteBatchId) == 0 {
				response.RespOKFail(writer, response.FileNotDeleted, "此文件未被删除")
				return err
			}
			deleteBatchIds[i] = urs[i].DeleteBatchId
		}
		// 软删除这些delete_batch_id记录
		err = models.DB.Where("delete_batch_id IN", deleteBatchIds).
			Delete(&models.RecoveryBatch{}).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}
		response.RespOKSuccess(writer, response.Success, nil, "删除成功")
		return nil
	})

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		for _, userFileId := range userFileIdList {
			if err := models.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, userFileId).
				Delete(&models.RecoveryBatch{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.RespOK(writer, 0, false, nil, "清空回收站失败")
		return
	}
	response.RespOkWithDataList(writer, 0, nil, 0, "清空回收站成功")
}

func RestoreFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub, isExist := models.GetUserBasicFromContext(c)
	if !isExist {
		response.RespUnAuthorized(writer)
		return
	}
	// 绑定请求参数
	var req api.RecoverFileReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.RespBadReq(writer, "参数错误")
		return
	}

	var urs []models.UserRepository

	models.DB.Transaction(func(tx *gorm.DB) error {
		// 找到recovery_batch中的记录
		var rb models.RecoveryBatch
		err = tx.Where("delete_batch_id = ?", req.DeleteBatchNum).First(&rb).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}

		err = tx.Unscoped().Where("user_id = ? AND delete_batch_id = ?", ub.UserId, req.DeleteBatchNum).Find(&urs).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}
		if len(urs) == 0 {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return errors.New("错误不存在")
		}

		folder, err := models.FindFolderFromAbsPath(tx, ub.UserId, req.FilePath)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 文件不存在
		} else if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}
		// 文件夹存在
		for i := range urs {
			if urs[i].UserFileId == rb.UserFileId {
				// 找到删除的代表文件
				if urs[i].ParentId != folder.UserFileId {
					// 如果父文件不是folder，就变成是
					urs[i].ParentId = folder.UserFileId
				}
			}
			urs[i].DeletedAt = 0
			urs[i].DeleteBatchId = ""
		}
		err = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_file_id"}},                                          // 主键id
			DoUpdates: clause.AssignmentColumns([]string{"parent_id", "delete_batch_id", "deleted_at"}), // 要更新的列名
		}).Create(urs).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}

		// 最后，删除回收站记录
		err = tx.Where("delete_batch_id = ?", req.DeleteBatchNum).Delete(&models.RecoveryBatch{}).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}
		response.RespOKSuccess(writer, response.Success, nil, "还原成功")
		return nil
	})
}
