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
