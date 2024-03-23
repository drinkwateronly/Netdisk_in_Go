package file_service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"strings"
)

// RenameFile
// @Summary 文件重命名
// @Tags Files
// @Accept json
// @Produce json
// @Param req query api.MoveFileReqAPI true "请求"
// @Success 200 {object} response.RespData "响应"
// @Router /file/getfilelist [GET]
func RenameFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	var req api.RenameFileRequest
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespBadReq(writer, "出现错误")
		return
	}
	// 校验文件名称
	if strings.ContainsAny(req.FileName, "|<>/\\:*?\"") {
		response.RespOKFail(writer, response.FileNameNotValid, "命名失败，文件名称出现非法字符")
		return
	}
	// 更新文件名
	res := models.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, req.UserFileId).
		Updates(models.UserRepository{
			FileName: req.FileName,
		})
	if models.IsDuplicateEntryErr(res.Error) {
		// 由于user_repository表中设置了UNIQUE INDEX `user_id`(`user_id`, `parent_id`, `file_name`, `extend_name`, `file_type`)
		// 因此文件记录的上述字段重复时会触发Error 1062 (23000): Duplicate entry
		response.RespOKFail(writer, response.FileNameNotValid, "文件名重复")
		return
	} else if res.RowsAffected == 0 {
		// 没有出现重复，update也没有影响行数
		response.RespOKFail(writer, response.FileNotExist, "文件不存在")
		return
	}
	if err != nil {
		response.RespOKFail(writer, response.DatabaseError, "数据库出错")
		return
	}
	response.RespOKSuccess(writer, 0, nil, "文件名修改成功")
	return
}

// MoveFile
// @Summary 文件移动
// @Tags Files
// @Accept json
// @Produce json
// @Param req query api.MoveFileReqAPI true "请求"
// @Success 200 {object} api.RespData
// @Failure default {object} api.RespData
// @Router /file/getfilelist [GET]
func MoveFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	// 绑定请求参数
	var req api.MoveFileReqAPI
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespBadReq(writer, "出现错误")
		return
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 查询源文件是否存在
		sourceFileUr, isExist := models.FindUserFileById(tx, ub.UserId, req.UserFileId)
		if !isExist {
			return errors.New("源文件夹不存在")
		}
		// 根据req中目的文件夹的绝对路径，查询该目的文件夹是否存在
		destFolder, err := models.FindFolderFromAbsPath(tx, ub.UserId, req.FilePath)
		if err != nil || destFolder == nil {
			return errors.New("目的文件夹不存在")
		}
		if sourceFileUr.ParentId == destFolder.UserFileId {
			return errors.New("该文件已在当前目录中")
		}
		// 源文件是否是文件夹
		if sourceFileUr.IsDir == 0 {
			// 移动的是文件
			// 查询是否有同名文件的存在，有则重命名源文件
			res := models.DB.
				Where("user_id = ? AND parent_id = ? AND file_name = ? AND extend_name = ?",
					ub.UserId, destFolder.UserFileId, sourceFileUr.FileName, sourceFileUr.ExtendName).
				Find(&models.UserRepository{})
			if res.Error != nil {
				return err
			}
			fileName := sourceFileUr.FileName
			// 有同名文件，则重新命名，添加后缀
			if res.RowsAffected != 0 {
				fileName = filehandler.RenameConflictFile(fileName)
			}
			// 更新源文件记录
			if err := models.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, req.UserFileId).
				Updates(&models.UserRepository{
					FilePath: req.FilePath,          // 新路径
					ParentId: destFolder.UserFileId, // 新父文件id
					FileName: fileName,              // 文件名
				}).Error; err != nil {
				return err
			}
			return nil
		}
		// 移动的是文件夹，即ur.Isdir == 1
		// 移动文件夹时嵌套文件夹为非法操作
		// 例如源文件夹'/A/B'移动到目的文件夹`A/B/C`中是非法的，因为C被B包含
		sourcePath := req.FilePath
		destPath := filehandler.ConCatFileFullPath(sourceFileUr.FilePath, sourceFileUr.FileName)
		// 从路径名判断，源文件夹是否被目的文件夹包含
		if req.FilePath != "/" && strings.HasPrefix(destPath, sourcePath) {
			return errors.New("目的文件夹在所移动文件夹内")
		}

		var allFiles []models.UserRepository
		// 找到源文件夹下所有文件记录，包括源文件夹本身
		err = models.DB.Raw(`with RECURSIVE temp as
(
    SELECT * from user_repository where user_file_id = ?
    UNION ALL
    SELECT ur.* from user_repository as ur,temp t 
	where ur.parent_id=t.user_file_id AND ur.deleted_at is NULL
)
select * from temp;`, sourceFileUr.UserFileId).Find(&allFiles).Error
		if err != nil {
			return err
		}
		// 更新文件记录
		// todo:文件同名冲突处理
		// todo:文件夹同名冲突处理
		prePathLen := len(allFiles[0].FilePath)
		allFiles[0].ParentId = destFolder.UserFileId
		allFiles[0].FilePath = req.FilePath
		for i := 1; i < len(allFiles); i++ {
			// 将父文件名称替换
			if req.FilePath == "/" {
				allFiles[i].FilePath = allFiles[i].FilePath[prePathLen:]
			} else {
				allFiles[i].FilePath = req.FilePath + "/" + allFiles[i].FilePath[prePathLen:]
			}
		}
		return nil
	})
	if err != nil {
		response.RespOK(writer, 999999, false, nil, err.Error())
		return
	}
	response.RespOK(writer, 0, true, nil, "文件移动成功")
}
