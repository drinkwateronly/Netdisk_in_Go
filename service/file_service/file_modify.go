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
// @Tags file
// @Accept json
// @Produce json
// @Param req query api.MoveFileReq true "请求"
// @Success 200 {object} response.RespData "响应"
// @Router /file/renamefile [GET]
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
// @Tags file
// @Accept json
// @Produce json
// @Param req query api.MoveFileReq true "请求"
// @Success 200 {object} response.RespData  "响应"
// @Router /file/movefile [POST]
func MoveFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub, isExist := models.GetUserBasicFromContext(c)
	if !isExist {
		response.RespUnAuthorized(writer)
		return
	}
	// 绑定请求参数
	var req api.MoveFileReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
		return
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 查询源文件是否存在
		sourceFileUr, err := models.FindUserFileById(tx, ub.UserId, req.UserFileId)
		if err != nil {
			response.RespOKFail(writer, response.FileNotExist, "源文件不存在")
			return err
		}
		// 源文件之前的FilePath
		preSourcePath := sourceFileUr.FilePath
		preSouceName := sourceFileUr.FileName
		curSoucePath := req.FilePath

		// 根据req中目的文件夹的绝对路径，查询该目的文件夹是否存在
		destFolder, err := models.FindFolderFromAbsPath(tx, ub.UserId, req.FilePath)
		if err != nil {
			response.RespOKFail(writer, response.FileNotExist, "目的文件夹不存在")
			return err
		}
		// 源文件已在目的文件夹下，不需要移动
		if sourceFileUr.ParentId == destFolder.UserFileId {
			response.RespOKFail(writer, response.FileRepeat, "该文件已在当前目录中")
			return errors.New("file repeat")
		}

		// 查询目的文件夹是否有同名文件的存在
		for {
			res := tx.Where("user_id = ? AND parent_id = ? AND file_name = ? AND extend_name = ?",
				ub.UserId, destFolder.UserFileId, sourceFileUr.FileName, sourceFileUr.ExtendName).
				Find(&models.UserRepository{})
			if res.Error != nil {
				response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
				return err
			}
			// 有同名文件，则重新命名，添加后缀
			if res.RowsAffected != 0 {
				sourceFileUr.FileName = filehandler.RenameConflictFile(sourceFileUr.FileName)
			} else {
				// 直到没有同名文件
				break
			}
		}
		sourceFileUr.ParentId = destFolder.UserFileId
		sourceFileUr.FilePath = req.FilePath

		if sourceFileUr.IsDir == 0 { // 源文件类型不是文件夹
			// 更新源文件记录
			err := models.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, req.UserFileId).
				Updates(&sourceFileUr).Error
			if err != nil {
				response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
				return err
			}
			response.RespOK(writer, 0, true, nil, "文件移动成功")
			return nil
		}
		// 源文件类型是文件夹
		var allFiles []*models.UserRepository
		// 找到源文件夹下所有文件记录，包括源文件夹本身
		allFiles, err = models.FindAllFilesFromFileId(tx, ub.UserId, req.UserFileId)
		if err != nil {
			response.RespOKFail(writer, response.FileNotExist, "源文件夹不存在")
			return err
		}

		// allFiles[0]变为sourceFileUr，此处为了后续一次性将allFiles更新到数据库
		// 而不需要先更新sourceFileUr，在更新剩余的allFiles[1:]
		allFiles[0].ParentId = sourceFileUr.ParentId
		allFiles[0].FileName = sourceFileUr.FileName
		allFiles[0].FilePath = sourceFileUr.FilePath

		// 新的路径前缀
		var newPathPrefix string
		if req.FilePath == "/" {
			// curSoucePath就是"/"
			newPathPrefix = "/" + allFiles[0].FileName
		} else {
			newPathPrefix = curSoucePath + "/" + allFiles[0].FileName
		}
		// 之前的前缀长度
		prePrefixLen := len(filehandler.ConCatFileFullPath(preSourcePath, preSouceName))

		// 移动文件，开始更新文件记录
		for i := 1; i < len(allFiles); i++ {
			if allFiles[i].UserFileId == destFolder.UserFileId {
				// 目的文件夹在源文件夹中，即源文件夹包括目的文件夹，会导致文件夹无限嵌套
				response.RespOKFail(writer, response.FolderLoopError, "非法操作，源文件夹包括目的文件夹")
				return errors.New("folder loop error")
			}
			// 将旧的路径前缀换为新的路径前缀
			allFiles[i].FilePath = newPathPrefix + allFiles[i].FilePath[prePrefixLen:]
		}
		// 直接更新这些文件记录
		err = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_file_id"}},                                   // 主键id
			DoUpdates: clause.AssignmentColumns([]string{"parent_id", "file_path", "file_name"}), // 要更新的列名
		}).Create(allFiles).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}
		response.RespOK(writer, 0, true, nil, "文件移动成功")
		return nil
	})
}

func MoveFileInBatch(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub, isExist := models.GetUserBasicFromContext(c)
	if !isExist {
		response.RespUnAuthorized(writer)
		return
	}
	// 绑定请求参数
	var req api.MoveFileInBatchReq
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

	// 要更新的所有文件夹记录
	var allFilesForUpdate []*models.UserRepository

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 循环所有要移动的源文件id
		for _, userFileId := range userFileIdList {
			// 查询源文件是否存在
			sourceFileUr, err := models.FindUserFileById(tx, ub.UserId, userFileId)
			if err != nil {
				response.RespOKFail(writer, response.FileNotExist, "源文件不存在")
				return err
			}
			// 源文件之前的FilePath
			preSourcePath := sourceFileUr.FilePath
			preSouceName := sourceFileUr.FileName
			curSoucePath := req.FilePath

			// 根据req中目的文件夹的绝对路径，查询该目的文件夹是否存在
			destFolder, err := models.FindFolderFromAbsPath(tx, ub.UserId, req.FilePath)
			if err != nil {
				response.RespOKFail(writer, response.FileNotExist, "目的文件夹不存在")
				return err
			}
			// 源文件已在目的文件夹下，不需要移动
			if sourceFileUr.ParentId == destFolder.UserFileId {
				response.RespOKFail(writer, response.FileRepeat, "该文件已在当前目录中")
				return errors.New("file repeat")
			}
			// 查询目的文件夹是否有同名文件的存在
			for {
				res := tx.Where("user_id = ? AND parent_id = ? AND file_name = ? AND extend_name = ?",
					ub.UserId, destFolder.UserFileId, sourceFileUr.FileName, sourceFileUr.ExtendName).
					Find(&models.UserRepository{})
				if res.Error != nil {
					response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
					return err
				}
				// 有同名文件，则重新命名，添加后缀
				if res.RowsAffected != 0 {
					sourceFileUr.FileName = filehandler.RenameConflictFile(sourceFileUr.FileName)
				} else {
					// 直到没有同名文件
					break
				}
			}
			sourceFileUr.ParentId = destFolder.UserFileId
			sourceFileUr.FilePath = req.FilePath

			if sourceFileUr.IsDir == 0 { // 源文件类型不是文件夹
				// 更新源文件记录
				allFilesForUpdate = append(allFilesForUpdate, sourceFileUr)
				continue
			}

			// 找到源文件夹下所有文件记录，包括源文件夹本身
			allFiles, err := models.FindAllFilesFromFileId(tx, ub.UserId, userFileId)
			if err != nil {
				response.RespOKFail(writer, response.FileNotExist, "源文件夹不存在")
				return err
			}

			// allFiles[0]变为sourceFileUr，此处为了后续一次性将allFiles更新到数据库
			// 而不需要先更新sourceFileUr，在更新剩余的allFiles[1:]
			allFiles[0].ParentId = sourceFileUr.ParentId
			allFiles[0].FileName = sourceFileUr.FileName
			allFiles[0].FilePath = sourceFileUr.FilePath

			// 新的路径前缀
			var newPathPrefix string
			if req.FilePath == "/" {
				// curSoucePath就是"/"
				newPathPrefix = "/" + allFiles[0].FileName
			} else {
				newPathPrefix = curSoucePath + "/" + allFiles[0].FileName
			}
			// 之前的前缀长度
			prePrefixLen := len(filehandler.ConCatFileFullPath(preSourcePath, preSouceName))

			// 移动文件，开始更新文件记录
			for i := 1; i < len(allFiles); i++ {
				if allFiles[i].UserFileId == destFolder.UserFileId {
					// 目的文件夹在源文件夹中，即源文件夹包括目的文件夹，会导致文件夹无限嵌套
					response.RespOKFail(writer, response.FolderLoopError, "非法操作，源文件夹包括目的文件夹")
					return errors.New("folder loop error")
				}
				// 将旧的路径前缀换为新的路径前缀
				allFiles[i].FilePath = newPathPrefix + allFiles[i].FilePath[prePrefixLen:]
			}
			allFilesForUpdate = append(allFilesForUpdate, allFiles...)
		}
		// 直接更新这些文件记录
		err = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_file_id"}},                                   // 主键id
			DoUpdates: clause.AssignmentColumns([]string{"parent_id", "file_path", "file_name"}), // 要更新的列名
		}).Create(allFilesForUpdate).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}
		response.RespOK(writer, 0, true, nil, "文件移动成功")
		return nil
	})
}
