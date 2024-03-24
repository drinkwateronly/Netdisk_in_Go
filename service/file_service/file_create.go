package file_service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"os"
	"time"
)

// CreateFile
// @Summary 文件创建
// @Description 仅支持excel，word，ppt文件的创建
// @Accept json
// @Produce json
// @Param req body api.CreateFileReq true "请求"
// @Success 200 {object} response.RespData "响应"
// @Router /createFile [POST]
func CreateFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	// 绑定请求参数
	var req api.CreateFileReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数不正确")
		return
	}
	// 前端仅支持excel，word，ppt的创建，且文件扩展名名全小写
	if req.ExtendName != "xlsx" && req.ExtendName != "docx" && req.ExtendName != "pptx" {
		response.RespOK(writer, response.NotSupport, false, nil, "文件类型不支持")
		return
	}
	// 开启事务
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 查询父文件夹记录
		parentDir, err := models.FindParentDirFromFilePath(tx, ub.UserId, req.FilePath)
		if err != nil {
			response.RespOK(writer, response.ParentNotExist, false, nil, "无法找到父文件夹")
			return err
		}
		// 开始创建文件
		userFileUUID := common.GenerateUUID()
		poolFileUUID := common.GenerateUUID()

		// 新增用户文件存储池记录
		err = tx.Create(&models.UserRepository{
			UserFileId: userFileUUID,
			UserId:     ub.UserId,
			FileId:     poolFileUUID,
			ParentId:   parentDir.UserFileId,
			FilePath:   req.FilePath,
			FileName:   req.FileName,
			FileType:   filehandler.EDITABLE, // 支持的三种文件类型都是EDITABLE
			IsDir:      0,
			ExtendName: req.ExtendName,
			ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
			UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
			FileSize:   0,
		}).Error
		if err != nil {
			// 因为user_repository表中将(`user_id`,`parent_id`,`file_name`,`extend_name`,`file_type`)作为唯一索引
			// 文件是否重复交由数据库是否返回错误代码是否为1062进行判断
			response.RespOK(writer, response.FileRepeat, false, nil, "该文件在当前文件夹已存在")
			return err
		}

		// 本地创建文件
		savePath := "./repository/upload_file/" + poolFileUUID
		file, err := os.Create(savePath)
		if err != nil {
			response.RespOK(writer, response.FileSaveError, false, nil, "文件保存出错")
			return err
		}
		err = file.Close()
		if err != nil {
			response.RespOK(writer, response.FileSaveError, false, nil, "文件保存出错")
			return err
		}

		// 新增中心存储池记录
		err = tx.Create(&models.RepositoryPool{
			FileId: poolFileUUID,
			Hash:   "d41d8cd98f00b204e9800998ecf8427e", // 创建文件时，文件大小为0，默认的哈希
			Size:   0,
			Path:   savePath,
		}).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "出错")
			return err
		}
		response.RespOK(writer, 0, true, nil, "创建文件成功")
		return nil
	})
}

// CreateFolder
// @Summary 文件夹创建
// @Accept json
// @Produce json
// @Param req body api.CreateFolderReq true "请求"
// @Success 200 {object} response.RespData "响应"
// @Router /createFold [POST]
func CreateFolder(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	var r api.CreateFolderReq
	err := c.ShouldBind(&r)
	if err != nil {
		response.RespBadReq(writer, "请求参数不正确")
		return
	}
	// 开启事务
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 查询父文件夹记录
		parentDir, err := models.FindParentDirFromFilePath(tx, ub.UserId, r.FolderPath)
		if err != nil {
			response.RespOKFail(writer, response.ParentNotExist, "父文件夹不存在")
			return err
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
				response.RespOK(writer, api.FileRepeat, false, nil, "同名文件夹已存在")
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
			response.RespOKFail(writer, response.FileRepeat, "当前目录下文件夹已存在")
			return err
		}
		response.RespOK(writer, response.Success, true, nil, "创建文件夹成功")
		return nil
	})
}
