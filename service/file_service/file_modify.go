package file_service

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"strings"
)

// RenameFile
// @Summary 文件重命名
// @Description 实现了文件重命名的接口
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
		if ubs[0].FileName == req.FileName {
			response.RespOKSuccess(writer, 0, nil, "文件名修改成功")
			return nil
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
		// user_file_id重复时更新，不存在时插入，此处是为了多个记录的更新，不会插入新数据。
		res := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_file_id"}},                      // 主键id
			DoUpdates: clause.AssignmentColumns([]string{"file_name", "file_path"}), // 要更新的列名
		}).Create(ubs)

		if models.IsDuplicateEntryErr(res.Error) {
			response.RespOKFail(writer, response.DatabaseError, "文件名重复")
			return err
		}
		if res.Error != nil {
			response.RespOKFail(writer, response.DatabaseError, "数据库出错")
			return err
		}
		if res.RowsAffected == 0 {
			// update也没有影响行数，即文件记录不存在
			response.RespOKFail(writer, response.FileNotExist, "文件不存在")
			return err
		}
		response.RespOKSuccess(writer, 0, nil, "文件名修改成功")
		return nil
	})
}

// MoveFileRepost
// @Summary 文件移动-转发请求版
// @Description 将文件移动的请求转化为文件批量移动的请求，交由文件批量移动接口处理，减少冗余代码
// @Tags file
// @Accept json
// @Produce json
// @Param req body api.MoveFileReq true "请求"
// @Success 200 {object} response.RespData  "响应"
// @Router /file/movefile [POST]
func MoveFileRepost(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	_, isExist := models.GetUserBasicFromContext(c)
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
	// 将文件转发请求转发到批量文件转发请求
	c.Request.URL.Path = "/file/batchmovefile"
	// 更改请求参数
	newReqByte, err := json.Marshal(api.MoveFileInBatchReq{
		UserFileIds: req.UserFileId,
		FilePath:    req.FilePath,
	})
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
		return
	}
	// 新请求参数重新放到body中
	c.Request.Body = io.NopCloser(bytes.NewReader(newReqByte))
}

// MoveFileInBatch
// @Summary 文件批量移动
// @Description 实现了文件批量移动的接口
// @Tags file
// @Accept json
// @Produce json
// @Param req body api.MoveFileInBatchReq true "请求"
// @Success 200 {object} response.RespData  "响应"
// @Router /file/batchmovefile [POST]
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
			// 源文件的FilePath
			preSourcePath := sourceFileUr.FilePath
			preSourceName := sourceFileUr.FileName
			curSourcePath := req.FilePath

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
			if sourceFileUr.UserFileId == destFolder.UserFileId {
				response.RespOKFail(writer, response.FileRepeat, "非法操作，源文件夹包括目的文件夹")
				return errors.New("file repeat")
			}

			// 如果目的文件夹下有和源文件同名的文件，则需要重命名源文件
			for {
				// 此处使用到了多列联合索引
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

			// 源文件类型不是文件夹
			if sourceFileUr.IsDir == 0 {
				// 源文件类型是文件夹，添加到
				allFilesForUpdate = append(allFilesForUpdate, sourceFileUr)
				continue
			} else {
				// 源文件类型是文件夹
				var allFiles []*models.UserRepository
				// 找到源文件夹下所有文件记录，包括源文件夹本身
				allFiles, err = models.FindAllFilesFromFileId(tx, ub.UserId, userFileId)
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
					newPathPrefix = curSourcePath + "/" + allFiles[0].FileName
				}
				// 之前的前缀长度
				prePrefixLen := len(filehandler.ConCatFileFullPath(preSourcePath, preSourceName))

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

// MoveFile
// @Summary 文件移动-实现版本
// @Description 实现了的文件移动的接口
// @Tags file
// @Accept json
// @Produce json
// @Param req body api.MoveFileReq true "请求"
// @Success 200 {object} response.RespData  "响应"
// @Router /file/movefile [POST]
func MoveFile(c *gin.Context) {
	/*
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
			// 源文件的FilePath
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
			if sourceFileUr.UserFileId == destFolder.UserFileId {
				response.RespOKFail(writer, response.FileRepeat, "非法操作，源文件夹包括目的文件夹")
				return errors.New("file repeat")
			}

			// 如果目的文件夹下有和源文件同名的文件，则需要重命名源文件
			for {
				// 此处使用到了多列联合索引
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

			// 源文件类型不是文件夹
			if sourceFileUr.IsDir == 0 {
				// 直接更新源文件记录即可
				err := models.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, req.UserFileId).
					Updates(&sourceFileUr).Error
				if err != nil {
					response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
					return err
				}
				response.RespOK(writer, 0, true, nil, "文件移动成功")
				return nil // return
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
	*/
}

// CopyFile
// @Summary 文件复制
// @Description 实现了的单个文件或文件夹复制的接口
// @Tags file
// @Accept json
// @Produce json
// @Param req body api.CopyFileReq true "请求"
// @Success 200 {object} response.RespData  "响应"
// @Router /file/copyfile [POST]
func CopyFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub, isExist := models.GetUserBasicFromContext(c)
	if !isExist {
		response.RespUnAuthorized(writer)
		return
	}
	// 绑定请求参数
	var req api.CopyFileReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
		return
	}
	// 用于记录复制文件的总大小
	var totalCopySize uint64
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 源文件类型是文件夹
		var allFiles []*models.UserRepository
		// 如果复制的文件是文件夹，则会找到该文件夹记录allFiles[0]及内部所有文件记录allFiles[1:]
		// 如果复制的文件是文件夹，则会找到该文件记录allFiles[0]
		allFiles, err = models.FindAllFilesFromFileId(tx, ub.UserId, req.UserFileId)
		if err != nil {
			response.RespOKFail(writer, response.FileNotExist, "源文件不存在")
			return err
		}
		// 记录源文件复制前的信息
		preSourceId := allFiles[0].UserFileId
		preSourcePath := allFiles[0].FilePath
		preSourceName := allFiles[0].FileName
		curSourcePath := req.FilePath

		// 根据req中目的文件夹的绝对路径，查询该目的文件夹是否存在
		destFolder, err := models.FindFolderFromAbsPath(tx, ub.UserId, req.FilePath)
		if err != nil {
			response.RespOKFail(writer, response.FileNotExist, "目的文件夹不存在")
			return err
		}

		// 如果目的文件夹下有和源文件同名的文件，则需要重命名源文件
		for {
			// 此处使用到了多列联合索引
			res := tx.Where("user_id = ? AND parent_id = ? AND file_name = ? AND extend_name = ?",
				ub.UserId, destFolder.UserFileId, allFiles[0].FileName, allFiles[0].ExtendName).
				Find(&models.UserRepository{})
			if res.Error != nil {
				response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
				return err
			}
			// 有同名文件，则重新命名，添加后缀
			if res.RowsAffected != 0 {
				allFiles[0].FileName = filehandler.RenameConflictFile(allFiles[0].FileName)
			} else {
				// 直到没有同名文件
				break
			}
		}

		allFiles[0].UserFileId = common.GenerateUUID()
		allFiles[0].ParentId = destFolder.UserFileId
		allFiles[0].FilePath = req.FilePath
		totalCopySize += allFiles[0].FileSize

		// 源文件类型是文件夹
		if allFiles[0].IsDir == 1 {
			uuidMap := make(map[string]string)
			// 旧和新的用户文件id映射
			uuidMap[preSourceId] = allFiles[0].UserFileId

			// 新的路径前缀
			var newPathPrefix string
			if req.FilePath == "/" {
				// curSoucePath就是"/"
				newPathPrefix = "/" + allFiles[0].FileName
			} else {
				newPathPrefix = curSourcePath + "/" + allFiles[0].FileName
			}
			// 之前的前缀长度
			prePrefixLen := len(filehandler.ConCatFileFullPath(preSourcePath, preSourceName))

			// 移动文件，开始更新文件记录
			for i := 1; i < len(allFiles); i++ {
				// 获取当前文件的父文件夹的新uuid
				newParentUUID, ok := uuidMap[preSourceId]
				newUUID := common.GenerateUUID()
				if ok {
					// 父文件新uuid存在
					allFiles[i].ParentId = newParentUUID
				} else {
					// 父文件新uuid不存在，则生成一个，并加入到map中
					uuidMap[allFiles[i].ParentId] = common.GenerateUUID()
					allFiles[i].ParentId = newParentUUID
				}
				if allFiles[i].IsDir == 1 {
					// 当前文件是文件夹，生成一个uuid
					uuidMap[allFiles[i].UserFileId] = newUUID
				}
				// 文件新uuid
				allFiles[i].UserFileId = newUUID
				// 文件新路径
				allFiles[i].FilePath = newPathPrefix + allFiles[i].FilePath[prePrefixLen:]
				// 总共复制的文件大小
				totalCopySize += allFiles[i].FileSize
			}
		}

		// 计算复制后容量是否超过了用户容量
		ub.StorageSize += totalCopySize
		if ub.TotalStorageSize < ub.StorageSize {
			response.RespOKFail(writer, response.StorageNotEnough, "网盘空间不足")
			return err
		}

		// 创建新文件记录
		err = tx.Create(allFiles).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}

		// 更新用户记录
		err = tx.Where("user_id = ?", ub.UserId).Updates(&ub).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
			return err
		}

		// 完成
		response.RespOKSuccess(writer, response.Success, nil, "文件移动成功")
		return nil
	})
}
