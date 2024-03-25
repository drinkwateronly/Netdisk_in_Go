package file_service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	var req api.RenameFileReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
		return
	}
	// 校验文件名称
	if strings.ContainsAny(req.FileName, "|<>/\\:*?\"\n\t\r") {
		response.RespOKFail(writer, response.FileNameNotValid, "文件名称出现非法字符")
		return
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 当UserFileId对应文件时，FindAllFilesFromFileId只会找到文件本身的记录
		// 当UserFileId对应文件夹时，FindAllFilesFromFileId会找到该文件夹及其内部所有文件的记录
		ubs, err := models.FindAllFilesFromFileId(tx, ub.UserId, req.UserFileId)
		if err != nil {
			return err
		}

		// 根据递归查询的结果来看，ubs[0]是要更名的文件

		// ubs[0]如果是文件夹，且其内部有文件，则要修改其内部文件的路径中对应该文件夹的名称
		// 例如，修改"/111/222"为"/111/333"时，该文件夹内部的所有文件路径都要改为"/111/333/..."
		// 这个过程就是拼接 "/111/" + 新修改文件名 + "/..."，因此使用双指针分割字符串
		var left, right int
		if ubs[0].FilePath == "/" {
			left = 1 // 例外，len("/")
		} else {
			left = len(ubs[0].FilePath + "/") // 定位到路径中修改文件夹的左，即len("/111/")
		}
		curFolderFullPath := filehandler.ConCatFileFullPath(ubs[0].FilePath, ubs[0].FileName)
		right = len(curFolderFullPath) // 定位到路径中修改文件夹的右，即len("/111/333")

		// 修改剩下文件的路径信息
		for i := 1; i < len(ubs); i++ {
			tmpPath := ubs[i].FilePath
			if len(tmpPath) < right {
				// 定位右侧超过了路径长度，所以路径右侧是空的，就不需要拼接右边
				ubs[i].FilePath = tmpPath[:left] + req.FileName
			} else {
				ubs[i].FilePath = tmpPath[:left] + req.FileName + tmpPath[right:]
			}
		}
		// 最后再修改该文件的文件名
		ubs[0].FileName = req.FileName
		// user_file_id重复时更新，不重复时插入，此处是为了多个记录的更新，不会插入新数据。
		res := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_file_id"}},                      // 主键id
			DoUpdates: clause.AssignmentColumns([]string{"file_name", "file_path"}), // 要更新的列名
		}).Create(ubs)

		if res.RowsAffected == 0 {
			// update也没有影响行数，即文件记录不存在
			response.RespOKFail(writer, response.FileNotExist, "文件不存在")
			return err
		}
		if res.Error != nil {
			response.RespOKFail(writer, response.DatabaseError, "数据库出错")
			return err
		}
		response.RespOKSuccess(writer, 0, nil, "文件名修改成功")
		return nil
	})
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
