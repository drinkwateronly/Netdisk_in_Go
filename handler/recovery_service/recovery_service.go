package recovery_service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"strings"
	"time"
)

// GetRecoveryFileList
// @Summary 获取回收站文件列表
// @Description
// @Tags recovery
// @Accept json
// @Produce json
// @Success 200 {object} response.RespDataList{datalist=[]api.RecoveryListResp} "响应"
// @Router /recoveryfile/list [GET]
func GetRecoveryFileList(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
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
// @Description
// @Tags recovery
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
// @Description
// @Tags recovery
// @Accept json
// @Produce json
// @Param userFileId body api.DelRecoveryInBatchReq true "请求"
// @Success 200 {object} response.RespData "响应"
// @Router /recoveryfile/batchdelete [POST]
func DelRecoveryInBatch(c *gin.Context) {
	writer := c.Writer
	ub, isExist := models.GetUserBasicFromContext(c)
	if !isExist {
		response.RespUnAuthorized(writer)
		return
	}
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
		err := tx.Unscoped().Where("user_id = ? AND user_file_id IN ? AND deleted_at <> 0", ub.UserId, userFileIdList).
			Find(&urs).Error
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
		err = models.DB.Where("delete_batch_id IN ?", deleteBatchIds).
			Delete(&models.RecoveryBatch{}).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}
		response.RespOKSuccess(writer, response.Success, nil, "删除成功")
		return nil
	})

}

// RestoreFile
// @Summary 回收站文件恢复
// @Description
// @Tags file
// @Accept json
// @Produce json
// @Param req body api.RecoverFileReq true "请求"
// @Success 200 {object} response.RespData  "响应"
// @Router /recoveryfile/restorefile [POST]
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
	// 用于存放要被恢复的文件记录
	var urs []models.UserRepository
	// 用于存放恢复的文件大小
	var totalRestoreSize uint64
	// 开启事务
	models.DB.Transaction(func(tx *gorm.DB) error {
		// 找到recovery_batch中的记录
		var rb models.RecoveryBatch
		err = tx.Where("delete_batch_id = ?", req.DeleteBatchNum).First(&rb).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}

		// 根据delete_batch_id找到user_repository中删除的文件记录
		err = tx.Unscoped(). // Unscoped表示查询软删除的数据
					Where("user_id = ? AND delete_batch_id = ?", ub.UserId, req.DeleteBatchNum).Find(&urs).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}
		if urs == nil {
			response.RespOKFail(writer, response.FileNotExist, "文件不存在")
			return errors.New("delete file not exist")
		}

		// 查询恢复文件的原路径是否存在，不存在则需要创建文件夹
		destFolder, err := models.FindFolderFromAbsPath(tx, ub.UserId, req.FilePath)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 原路径不存在，开始创建文件夹
			folderList := strings.Split(req.FilePath[1:], "/")
			curPath := "/"
			root, _ := models.FindRoot(tx, ub.UserId)
			curParentId := root.UserFileId
			var folder models.UserRepository
			for _, folderName := range folderList {
				// 当前路径curPath下有没有名为folderName的文件夹
				res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("user_id = ? AND file_name = ?  AND is_dir = 1 AND file_path = ?", ub.UserId, folderName, curPath).
					Find(&folder)
				// 文件夹不存在，创建名为folderName的文件夹记录
				if res.RowsAffected == 0 {
					folder = models.UserRepository{
						UserFileId: common.GenerateUUID(),
						UserId:     ub.UserId,
						FilePath:   curPath,
						FileName:   folderName,
						ParentId:   curParentId,
						FileType:   filehandler.DIRECTORY,
						IsDir:      1,
						ExtendName: "",
						ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
						UploadTime: time.Now().Format("2006-01-02 15:04:05"),
					}
					err = tx.Create(&folder).Error
					if err != nil {
						return err
					}
				}
				curPath += folderName
				curParentId = folder.UserFileId
			}
			// destFolder
			destFolder = &folder
		} else if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}

		// 恢复路径下是否已有重名文件，有时需要修改文件名
		preSourceName := rb.FileName
		for {
			// 此处使用到了多列联合索引
			res := tx.Where("user_id = ? AND parent_id = ? AND file_name = ? AND extend_name = ?",
				ub.UserId, destFolder.UserFileId, rb.FileName, rb.ExtendName).
				Find(&models.UserRepository{})
			if res.Error != nil {
				response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
				return err
			}
			// 有同名文件，则重新命名，添加后缀
			if res.RowsAffected != 0 {
				rb.FileName = filehandler.RenameConflictFile(rb.FileName)
			} else {
				break
			}
		}
		newSourceName := rb.FileName

		// 循环文件记录
		for i := range urs {
			totalRestoreSize += urs[i].FileSize
			if urs[i].UserFileId == rb.UserFileId {
				// 文件夹节点链接到目的文件夹下
				urs[i].ParentId = destFolder.UserFileId
				// 修改文件名
				urs[i].FileName = newSourceName
			} else {
				preFileFullPath := filehandler.ConCatFileFullPath(rb.FilePath, preSourceName)
				urs[i].FilePath = filehandler.ConCatFileFullPath(urs[i].FilePath[0:len(rb.FilePath)], newSourceName+urs[i].FilePath[len(preFileFullPath):])
			}
			urs[i].DeletedAt = 0
			urs[i].DeleteBatchId = ""
		}

		// 更新用户空间
		ub.StorageSize += totalRestoreSize
		if ub.TotalStorageSize < ub.StorageSize {
			response.RespOKFail(writer, response.StorageNotEnough, "网盘空间不足")
			return err
		}

		// 更新
		err = tx.Clauses(clause.OnConflict{ // 主键若重复则更新
			Columns:   []clause.Column{{Name: "user_file_id"}},                                                                    // 主键id
			DoUpdates: clause.AssignmentColumns([]string{"parent_id", "file_path", "file_name", "delete_batch_id", "deleted_at"}), // 要更新的列名
		}).Create(urs).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}

		// 最后，软删除回收站记录
		err = tx.Where("delete_batch_id = ?", req.DeleteBatchNum).Delete(&models.RecoveryBatch{}).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}
		response.RespOKSuccess(writer, response.Success, nil, "还原成功")
		return nil
	})
}
