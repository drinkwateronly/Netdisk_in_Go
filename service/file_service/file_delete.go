package file_service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"strings"
	"time"
)

// DeleteFile
// @Summary 文件单个删除接口
// @Tags file
// @Accept json
// @Produce json
// @Param req body api.DeleteFileReq true "请求"
// @Success 200 {object} response.RespData "响应"
// @Router /file/deletefile [POST]
func DeleteFile(c *gin.Context) {
	writer := c.Writer
	ub := c.MustGet("userBasic").(*models.UserBasic)
	// 绑定请求参数
	var req api.DeleteFileReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespBadReq(writer, "出现错误")
		return
	}
	// 批次id
	delBatchId := common.GenerateUUID()
	// 删除时间
	delTime := time.Now()
	// 本次删除文件的总大小
	var delStorageSize uint64
	// 开启事务，删除文件夹
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		ur, isExist := models.FindUserFileById(tx, ub.UserId, req.UserFileId)
		if !isExist {
			return errors.New("文件不存在")
		}
		// 当删除的是文件时，FindAllFilesFromFileId只会找到文件本身的记录
		// 当删除的是文件夹时，FindAllFilesFromFileId会找到该文件夹及其内部所有文件的记录
		ubs, err := models.FindAllFilesFromFileId(tx, ub.UserId, req.UserFileId)
		//err = models.DelAllFilesFromDir(tx, delBatchId, ub.UserId, ur.FilePath, ur.FileName)
		if err != nil {
			return err
		}
		// 软删除文件记录，返回删除文件总存储容量
		delStorageSize, err = models.SoftDelUserFiles(tx, delTime, delBatchId, ub.UserId, ubs...)
		if err != nil {
			return err
		}
		// 更新用户容量
		if err := models.UpdateUserStorageSize(tx, ub.UserId, ub.StorageSize-delStorageSize); err != nil {
			return err
		}
		// 将删除的文件信息添加到到回收站文件批次记录表，
		// 删除文件夹时，其下所有文件从属一个删除批次，因此只添加该文件夹的信息
		// 此时回收站只展示该文件夹，恢复该文件夹时将恢复同一批次的所有文件
		err = models.InsertToRecoveryBatch(tx, delTime, delBatchId, ur)
		return err
	})
	if err != nil {
		response.RespOK(writer, 0, false, nil, "删除文件失败")
		return
	}
	response.RespOKSuccess(writer, response.Success, nil, "删除成功")
	return
}

// DeleteFilesInBatch
// @Summary 文件批量删除接口
// @Tags file
// @Accept json
// @Produce json
// @Param req body api.DeleteFileInBatchReq true "请求"
// @Success 200 {object} response.RespData "响应"
// @Router /file/batchdeletefile [POST]
func DeleteFilesInBatch(c *gin.Context) {
	writer := c.Writer
	ub := c.MustGet("userBasic").(*models.UserBasic)
	// 绑定请求参数
	var req api.DeleteFileInBatchReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
		return
	}
	userFileIdList := strings.Split(req.UserFileIds, ",")
	if len(userFileIdList) == 0 {
		response.RespBadReq(writer, "请求参数错误")
		return
	}
	// 删除时间
	delTime := time.Now()
	// 本次删除文件的总大小
	var delStorageSize uint64
	// 开启事务，删除文件
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 找出这些文件信息
		urs, isExist := models.FindUserFilesByIds(tx, ub.UserId, userFileIdList)
		if !isExist {
			return errors.New("文件不存在")
		}
		for i := range urs {
			// 批次id
			delBatchId := common.GenerateUUID()
			// 当删除的是文件时，FindAllFilesFromFileId只会找到文件本身的记录
			// 当删除的是文件夹时，FindAllFilesFromFileId会找到该文件夹及其内部所有文件的记录
			ubs, err := models.FindAllFilesFromFileId(tx, ub.UserId, urs[i].UserFileId)
			if err != nil {
				return err
			}
			// 软删除文件记录，返回删除文件总存储容量
			delSize, err := models.SoftDelUserFiles(tx, delTime, delBatchId, ub.UserId, ubs...)
			delStorageSize += delSize
			if err != nil {
				return err
			}
			err = models.InsertToRecoveryBatch(tx, delTime, delBatchId, urs[i])
			if err != nil {
				return err
			}
		}
		// 更新用户容量
		return models.UpdateUserStorageSize(tx, ub.UserId, ub.StorageSize-delStorageSize)
	})
	if err != nil {
		response.RespOK(writer, 9999, false, nil, "删除文件失败")
		return
	}
	response.RespOKSuccess(writer, response.Success, nil, "删除成功")
	return
}
