package file_service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"os"
	"strings"
	"time"
)

// CreateFile
// @Summary 文件创建，仅支持excel，word，ppt的创建
// @Accept json
// @Produce json
// @Param req body api_models.CreateFileReqAPI true "请求"
// @Param username body string true "用户名"
// @Param password body string true "密码"
// @Success 200 {object} api_models.RespData{} ""
// @Failure 400 {object} string "参数出错"
// @Router /createFile [POST]
func CreateFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	// 绑定请求参数
	var req api.CreateFileReqAPI
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.RespBadReq(writer, "出现错误")
		return
	}
	// 仅支持excel，word，ppt的创建
	if req.ExtendName != "xlsx" && req.ExtendName != "docx" && req.ExtendName != "pptx" {
		response.RespOK(writer, response.FILETYPENOTSUPPORT, false, nil, "文件类型不支持")
		return
	}
	// 开启事务
	// todo:思考，事务A和B都进行在同一个路径创建同名文件，且在事务开启时都没找到同名文件，此时两个事务执行后是否会创建两个相同记录？
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 查询父文件夹记录
		parentDir, isExist, err := models.FindParentDirFromAbsPath(tx, ub.UserId, req.FilePath)
		if err != nil {
			response.RespOK(writer, response.DATABASEERROR, false, nil, "创建文件夹失败")
			return errors.New("database error" + err.Error())
		}
		if !isExist {
			response.RespOK(writer, response.PARENTNOTEXIST, false, nil, "无法找到父文件夹")
			return errors.New("parent directory not exist")
		}

		// 检查是否有重名文件
		_, isExist, err = models.FindFileByNameAndPath(tx, ub.UserId, req.FilePath, req.FileName, req.ExtendName)
		if err != nil {
			response.RespOK(writer, response.FILEREPEAT, false, nil, "文件在当前文件夹已存在")
			return errors.New("file repeat")
		}
		if isExist {
			response.RespOK(writer, response.FILEREPEAT, false, nil, "文件在当前文件夹已存在")
			return errors.New("file repeat")
		}
		// 创建文件
		userFileUUID := common.GenerateUUID()
		poolFileUUID := common.GenerateUUID()
		savePath := "./repository/upload_file/" + poolFileUUID
		file, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			response.RespOK(writer, response.FILECREATEERROR, false, nil, "文件保存出错")
			return err
		}
		err = file.Close()
		if err != nil {
			response.RespOK(writer, response.FILESAVEERROR, false, nil, "文件保存出错")
			return err
		}

		if err := tx.Create(&models.UserRepository{
			UserFileId: userFileUUID,
			UserId:     ub.UserId,
			FileId:     poolFileUUID,
			ParentId:   parentDir.UserFileId,
			FilePath:   req.FilePath,
			FileName:   req.FileName,
			FileType:   2, // 仅支持三种文件格式，因此类型是
			IsDir:      0,
			ExtendName: req.ExtendName,
			ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
			UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
			FileSize:   0,
		}).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.RepositoryPool{
			FileId: poolFileUUID,
			Hash:   "d41d8cd98f00b204e9800998ecf8427e", // 创建文件时，文件大小为0，默认的哈希
			Size:   0,
			Path:   savePath,
		}).Error; err != nil {
			return err
		}
		response.RespOK(writer, 0, true, nil, "创建文件成功")
		return nil
	})

}

// CreateFolder 文件上传
func CreateFolder(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	var r api.CreateFolderRequest
	err := c.ShouldBind(&r)
	if err != nil {
		response.RespBadReq(writer, "出现错误")
		return
	}
	// 开启事务
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 查询父文件夹记录
		parentDir, isExist, err := models.FindParentDirFromAbsPath(tx, ub.UserId, r.FolderPath)
		if err != nil {
			response.RespOK(writer, response.DATABASEERROR, false, nil, "创建文件夹失败")
			return errors.New("database error" + err.Error())
		}
		if !isExist {
			response.RespOK(writer, response.PARENTNOTEXIST, false, nil, "无法找到父文件夹")
			return errors.New("parent directory not exist")
		}
		// 不需要查询文件是否存在，因为user_repository表中将(`user_id`,`parent_id`,`file_name`,`extend_name`,`file_type`)作为唯一索引
		// 因此，文件是否重复交由数据库是否返回错误代码是否为1062进行判断
		/*
			// 查询父文件夹下同名文件夹
			res := tx.Where("user_id = ? AND file_name = ? AND file_path = ? AND file_type = ?", ub.UserId, r.FolderName, r.FolderPath, common.DIRECTORY).
				Find(&models.UserRepository{})
			if res.Error != nil {
				return res.Error
			}
			// 文件存在
			if res.RowsAffected != 0 {
				response.RespOK(writer, api.FILEREPEAT, false, nil, "同名文件夹已存在")
				return errors.New("file repeat")
			}
		*/
		// 新增文件记录
		err = tx.Create(&models.UserRepository{
			UserFileId: common.GenerateUUID(),
			UserId:     ub.UserId,
			FilePath:   r.FolderPath,
			ParentId:   parentDir.UserFileId,
			FileName:   r.FolderName,
			FileType:   filehandler.DIRECTORY,
			IsDir:      1,
			ExtendName: "",
			ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
			UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
		}).Error
		if err != nil {
			if models.IsDuplicateEntryErr(err) {
				response.RespOK(writer, response.FILEREPEAT, false, nil, "文件夹已存在")
				return err
			}
			response.RespOK(writer, response.DATABASEERROR, false, nil, "创建文件夹失败")
			return err
		}
		response.RespOK(writer, response.Success, true, nil, "创建文件夹成功")
		return nil
	})
}

func DeleteFile(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)

	type DeleteFileRequest struct {
		UserFileId string `json:"userFileId"`
	}
	var r DeleteFileRequest
	err := c.ShouldBind(&r)
	if err != nil {
		response.RespBadReq(writer, "出现错误")
		return
	}
	// 如果文件不存在，删除失败
	ur, isExist := models.FindUserFileById(models.DB, ub.UserId, r.UserFileId)
	if !isExist {
		// 找不到记录
		response.RespOK(writer, 1, false, nil, "文件不存在")
		return
	}
	// 开启事务，删除文件夹
	delBatchId := common.GenerateUUID()
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		if ur.FileType == filehandler.DIRECTORY { // 如果文件是文件夹
			// 递归进入文件夹，删除文件夹内部的文件
			err = models.DelAllFilesFromDir(delBatchId, ub.UserId, ur.FilePath, ur.FileName)
			if err != nil {
				return err
			}
			// 删除文件夹自己
			err = models.DB.Where("user_file_id = ?", ur.UserFileId).
				Delete(&models.UserRepository{}).Error
			if err != nil {
				return err
			}
		} else {
			err = models.DB.Where("user_id = ? and user_file_id = ?", ub.UserId, r.UserFileId).
				Updates(&models.UserRepository{}).Error
			if err != nil {
				return err
			}
		}
		// 添加到回收站
		err = models.AddFileToRecoveryBatch(ur, delBatchId)
		return err
	})
	if err != nil {
		response.RespOK(writer, 0, false, nil, "删除文件失败")
		return
	}
	response.RespOK(writer, 0, true, nil, "删除成功")
}

func DeleteFilesInBatch(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)

	type DeleteFilesRequest struct {
		UserFileIds string `json:"userFileIds"`
	}

	var r DeleteFilesRequest
	err := c.ShouldBind(&r)
	if err != nil {
		response.RespBadReq(writer, "出现错误")
		return
	}
	userFileIdList := strings.Split(r.UserFileIds, ",")
	// 开启事务，删除文件
	delBatchId := common.GenerateUUID()
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 找出这些文件信息
		var urList []*models.UserRepository
		err = models.DB.
			//Clauses(clause.Locking{Strength: "UPDATE"}). // 排他锁
			Where("user_id = ? and user_file_id in ?", ub.UserId, userFileIdList).
			Find(&urList).Error
		if err != nil {
			return err
		}
		// 循环文件
		for _, ur := range urList {
			if ur.FileType == filehandler.DIRECTORY {
				// 如果文件是文件夹
				// 递归进入文件夹，删除文件夹内部的文件
				err = models.DelAllFilesFromDir(delBatchId, ub.UserId, ur.FilePath, ur.FileName)
				if err != nil {
					return err
				}
				// 删除文件夹自己记录
				err = models.SoftDelUserFiles(delBatchId, ub.UserId, ur.UserFileId)
				if err != nil {
					return err
				}
			} else {
				// 是文件，直接删除
				err = models.SoftDelUserFiles(delBatchId, ub.UserId, ur.UserFileId)
				if err != nil {
					return err
				}
			}
			// 添加到回收站
			err = models.AddFileToRecoveryBatch(ur, delBatchId)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.RespOK(writer, 9999, false, nil, "删除文件失败")
		return
	}
	response.RespOK(writer, 0, true, nil, "删除成功")
}
