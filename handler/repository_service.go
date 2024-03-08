package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"net/http"
	ApiModels "netdisk_in_go/api_models"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
	"os"
	"strings"
	"time"
)

func FileUploadPrepare(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, err := utils.ParseCookieFromRequest(c)
	if err != nil {
		utils.RespOK(writer, 1001, false, nil, "cookie校验失败")
		return
	}
	// 绑定query请求参数
	var req ApiModels.FileUploadReqAPI
	err = c.ShouldBindQuery(&req)
	if err != nil {
		utils.RespBadReq(writer, "请求参数出错")
		return
	}

	// 处理出文件名、拓展名、文件的逻辑绝对路径、文件类型
	processedFileInfo := utils.GetFileInfoFromReq(req)

	// 开启事务
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		// 获取用户信息
		ub, isExist, err := models.FindUserByIdentity(tx, uc.UserId)
		if !isExist {
			return errors.New("用户不存在")
		}
		// 判断存储空间是否足够，前端已经做好了此判断工作。
		if ub.StorageSize+req.TotalSize > ub.TotalStorageSize {
			return errors.New("用户存储空间不足")
		}

		// 判断文件夹是否存在
		parentDir, isExist, err := models.FindParentDirFromAbsPath(tx, ub.UserId, req.FilePath)
		if err != nil {
			return errors.New("文件夹不存在")
		}

		// 判断文件在当前文件夹是否重名
		if _, isExist, err = models.FindFileByNameAndPath(tx, ub.UserId,
			processedFileInfo.AbsPath,
			processedFileInfo.FileName,
			processedFileInfo.ExtendName); isExist {
			return errors.New("文件在当前文件夹已存在")
		}
		if err != nil {
			return errors.New("文件夹不存在")
		}
		// 如果文件大小为0，则上传文件
		if req.TotalSize == 0 {
			utils.RespOK(writer, 0, true, gin.H{"skipUpload": false}, "开始上传文件")
			return nil
		}
		// 根据md5值判断文件在中心存储池是否已存在
		rp, isExist := models.FindFileByMD5(req.FileMD5)
		if !isExist { // 文件不存在，上传文件
			utils.RespOK(writer, 0, true, gin.H{"skipUpload": false}, "开始上传文件")
			return nil
		}

		// 到此处时，表示文件存在，应当进行文件秒传，只需要处理数据库即可，两种情况：
		// 		1.存放文件的文件夹不存在，需要创建文件夹记录
		// 		2.存放文件的文件夹存在，直接创建文件记录

		// 查存储文件的文件夹是否存在
		parentDir, isExist, err = models.FindParentDirFromAbsPath(tx, ub.UserId, processedFileInfo.AbsPath)
		if err != nil {
			return errors.New("文件夹不存在")
		}

		var parentId string     // 记录当前文件/文件夹的父文件夹id
		curPath := req.FilePath // 当前路径就是文件上传时候的根路径
		// if成立时，存放上传文件的文件夹不存在，这种情况常见于整个文件夹的上传时存在相对路径
		// 例在/123目录上传456/789/0.txt，接下来的步骤将在文件夹123按顺序创建文件夹456和789
		if !isExist {
			// 找到/123的文件id
			uploadRoot, _, err := models.FindParentDirFromAbsPath(tx, ub.UserId, curPath)
			if err != nil {
				return err
			}
			parentId = uploadRoot.UserFileId

			// 得到相对路径"456/789"
			var relativePath string

			relativePathLen := len(req.RelativePath) - len(req.FileFullName)
			relativePath = req.RelativePath[:relativePathLen-1]

			// 取出文件夹列表 [456, 789]，即文件相对路径先后进入的文件夹的列表
			folderList := strings.Split(relativePath, "/")
			// 接下来for循环中，进入curPath的文件夹，查询有没有folderName的文件夹，有则修改curPath进入下一级文件夹，无则创建文件夹folderName。
			for _, folderName := range folderList {
				var folder models.UserRepository
				// 当前文件上传的目录filePath有没有名为folderName的文件夹
				res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("user_id = ? AND file_name = ? AND file_path = ? AND is_dir = 1", ub.UserId, folderName, curPath).
					Find(&folder)
				if res.Error != nil {
					return res.Error
				}
				// 文件夹不存在，就创建在路径filePath的文件夹folderName
				if res.RowsAffected == 0 {
					folder = models.UserRepository{
						UserFileId: utils.GenerateUUID(),
						UserId:     ub.UserId,
						FilePath:   curPath,
						FileName:   folderName,
						ParentId:   parentId,
						FileType:   utils.DIRECTORY,
						IsDir:      1,
						ExtendName: "",
						ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
						UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
					}
					err = tx.Create(&folder).Error
					if err != nil {
						return err
					}
				}
				// 然后进入下一级目录，继续创建文件夹
				parentId = folder.UserFileId
				if curPath == "/" {
					curPath += folderName
				} else {
					curPath += "/" + folderName
				}
			}
		} else {
			// 文件的父文件夹存在，
			parentId = parentDir.UserFileId
		}
		// 文件夹创建完毕，开始创建文件
		ur := models.UserRepository{
			UserFileId: utils.GenerateUUID(),                     // 用户文件id
			UserId:     ub.UserId,                                // 用户id
			FileId:     rp.FileId,                                // 存储池文件id
			FilePath:   processedFileInfo.AbsPath,                //
			FileName:   processedFileInfo.FileName,               // 用户存储时的文件名
			ExtendName: processedFileInfo.ExtendName,             // 文件拓展名
			FileType:   processedFileInfo.FileType,               // 文件拓展名
			ParentId:   parentId,                                 // 父文件夹的id
			IsDir:      0,                                        // 是否是文件夹
			ModifyTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
			UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
			FileSize:   req.TotalSize,                            // 文件尺寸
		}

		if err := tx.Create(&ur).Error; err != nil {
			return err
		}
		newStorageSize := ub.StorageSize + req.TotalSize
		if err := tx.Where("user_id = ?", ub.UserId).
			Updates(models.UserBasic{StorageSize: newStorageSize}).Error; err != nil {
			return err
		}
		utils.RespOK(writer, 0, true, gin.H{"skipUpload": true}, "文件秒传")
		return nil
	})
	if err != nil {
		utils.RespOK(writer, 9999, false, nil, "文件上传失败："+err.Error())
	}
}

// FileUpload 文件上传
// FileUpload
// @Summary 文件上传
// @Produce json
// @Param req body ApiModels.FileUploadReqAPI true "文件上传请求"
// @Param cookie query string true "Cookie"
// @Success 200 {object} string "存储容量"
// @Failure 400 {object} string "参数出错"
// @Router /filetransfer/uploadfile [POST]
func FileUpload(c *gin.Context) {
	var savePath string
	var ur models.UserRepository
	var uc *utils.UserClaim
	var err error
	writer := c.Writer

	// 校验cookie
	var isAuth bool
	uc, isAuth = utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}

	// 绑定请求参数
	var req ApiModels.FileUploadReqAPI
	err = c.ShouldBind(&req)
	if err != nil {
		utils.RespOK(writer, 999999, false, nil, "参数出错")
		return
	}

	// 处理上传的文件分片
	chunkPath := fmt.Sprintf("./repository/chunk_file/%s-%d.chunk", req.FileMD5, req.ChunkNumber)
	//该分片可能之前上传过，例如上次上传时失败，此时略过上传
	//if !utils.IsChunkExist(chunkPath, req.CurrentChunkSize) {
	uploadedFile, err := c.FormFile("file")
	if err != nil || uploadedFile == nil {
		utils.RespOK(writer, 999999, false, nil, "上传出错")
		return
	}
	if err = c.SaveUploadedFile(uploadedFile, chunkPath); err != nil {
		utils.RespOK(writer, 999999, false, nil, "上传出错")
		return
	}
	//}
	// 如果不是最后一个分片，那么继续上传
	if req.ChunkNumber != req.TotalChunks {
		utils.RespOK(writer, 0, true, nil, "分片上传成功")
		return
	}

	// #############走到这里意味着最后一块分块上传完成，开始合并文件#############
	// 处理出文件名、拓展名、文件的逻辑绝对路径、文件类型
	processedFileInfo := utils.GetFileInfoFromReq(req)
	fmt.Printf("%q", processedFileInfo)

	// 所有文件分片上传完成，并得到了文件信息，开启事务
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		// 获取用户信息
		ub, isExist, err := models.FindUserByIdentity(tx, uc.UserId)
		if !isExist {
			utils.RespBadReq(writer, "用户不存在")
		}

		// 判断父文件夹是否存在
		//parentDir, _, err := models.FindParentDirFromAbsPath(tx, ub.UserId, req.FilePath)
		//if err != nil {
		//	return errors.New("文件夹不存在")
		//}

		// 判断文件在当前文件夹是否重名
		if _, isExist, err = models.FindFileByNameAndPath(tx, ub.UserId,
			processedFileInfo.AbsPath,
			processedFileInfo.FileName,
			processedFileInfo.ExtendName); isExist {
			return errors.New("文件在当前文件夹已存在")
		}
		if err != nil {
			return errors.New("文件夹不存在")
		}
		// 根据绝对路径 判断文件的父文件夹是否存在
		fmt.Println("processedFileInfo.AbsPath", processedFileInfo.AbsPath)

		parentDir, isExist, err := models.FindParentDirFromAbsPath(tx, ub.UserId, processedFileInfo.AbsPath)
		if err != nil {
			return err
		}

		var parentId string     // 记录当前文件/文件夹的父文件夹id
		curPath := req.FilePath // 当前路径就是文件上传时候的根路径
		// if成立时，存放上传文件的文件夹不存在，这种情况常见于整个文件夹的上传时存在相对路径
		// 例在/123目录上传456/789/0.txt，接下来的步骤将在文件夹123按顺序创建文件夹456和789
		if !isExist {
			// 找到/123的文件id
			uploadRoot, _, err := models.FindParentDirFromAbsPath(tx, ub.UserId, curPath)
			if err != nil {
				return err
			}
			parentId = uploadRoot.UserFileId

			// 得到相对路径"456/789"
			//strings.TrimSuffix(req.RelativePath, "/"+req.FileFullName)
			var relativePath string

			relativePathLen := len(req.RelativePath) - len(req.FileFullName)
			relativePath = req.RelativePath[:relativePathLen-1]

			// 取出文件夹列表 [456, 789]，即文件相对路径先后进入的文件夹的列表
			folderList := strings.Split(relativePath, "/")
			// 接下来for循环中，进入curPath的文件夹，查询有没有folderName的文件夹，有则修改curPath进入下一级文件夹，无则创建文件夹folderName。

			for _, folderName := range folderList {
				var folder models.UserRepository
				// 当前文件上传的目录filePath有没有名为folderName的文件夹
				res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("user_id = ? AND file_name = ? AND file_path = ? AND is_dir = 1", ub.UserId, folderName, curPath).
					Find(&folder)
				if res.Error != nil {
					return res.Error
				}
				// 文件夹不存在，就创建在路径filePath的文件夹folderName
				if res.RowsAffected == 0 {
					folder = models.UserRepository{
						UserFileId: utils.GenerateUUID(),
						UserId:     ub.UserId,
						FilePath:   curPath,
						FileName:   folderName,
						ParentId:   parentId,
						FileType:   utils.DIRECTORY,
						IsDir:      1,
						ExtendName: "",
						ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
						UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
					}
					err = tx.Create(&folder).Error
					if err != nil {
						return err
					}
				}
				// 新创建的文件夹id作为下一级文件的parentId
				parentId = folder.UserFileId
				// 然后进入下一级目录，继续创建文件夹
				if curPath == "/" {
					curPath += folderName
				} else {
					curPath += "/" + folderName
				}
			}
		} else {
			// 文件的父文件夹存在，
			parentId = parentDir.UserFileId
		}
		// 最后filePath变成所要上传的文件的绝对路径，上例中，则为/123/456/

		// 生成文件uuid
		poolFileId := utils.GenerateUUID()
		userFileId := utils.GenerateUUID()
		savePath = "./repository/upload_file/" + poolFileId

		// 将分块文件合并
		err = utils.MergeChunksToFile(req.FileMD5, poolFileId, req.TotalChunks)
		if err != nil {
			return err
		}

		// 校验md5
		mergeMD5, err := utils.GetFileMD5FromPath(savePath)
		if mergeMD5 != req.FileMD5 || err != nil {
			return err
		}

		// 开始写入数据库
		ur = models.UserRepository{
			UserFileId: userFileId, // 用户文件id
			FileId:     poolFileId, // 存储池文件id
			UserId:     ub.UserId,  // 用户id
			ParentId:   parentId,
			FilePath:   processedFileInfo.AbsPath,    //
			FileName:   processedFileInfo.FileName,   // 用户存储时的文件名
			ExtendName: processedFileInfo.ExtendName, // 文件拓展名
			FileType:   processedFileInfo.FileType,   // 文件类型
			IsDir:      0,
			FileSize:   req.TotalSize, // 文件大小
			ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
			UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
		}
		rp := models.RepositoryPool{
			FileId: poolFileId,
			Hash:   req.FileMD5,
			Size:   req.TotalSize,
			Path:   savePath,
		}
		// 插入文件记录repository_pool, user_repository，修改用户存储容量
		if err := tx.Create(&ur).Error; err != nil {
			return err
		}
		if err := tx.Create(&rp).Error; err != nil {
			return err
		}
		newStorageSize := ub.StorageSize + req.TotalSize
		if err := tx.Where("user_id = ?", ub.UserId).
			Updates(models.UserBasic{StorageSize: newStorageSize}).Error; err != nil {
			return err
		}
		return nil

	})
	if err != nil {
		utils.RespOK(writer, 9999, false, nil, "文件上传失败："+err.Error())
		return
	}

	utils.RespOK(writer, 0, true, nil, "文件上传成功")
	// resp后，用户已经收到文件上传的结果
	// 如果文件类型是图片/视频，则保存preview格式，方便后续前端预览
	switch ur.FileType {
	case utils.IMAGE:
		// 不处理错误
		_ = utils.SavePreviewFromImage(savePath, processedFileInfo.ExtendName)
	case utils.VIDEO:
		// 不处理错误
		_ = utils.SavePreviewFromVideo(savePath, 5)
	}
	// 开始删除分片文件，不删除会好点，后续可以统一删除
	//utils.DeleteAllChunks(fileMD5, totalChunks)
}

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
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, ApiModels.UNAUTHORIZED, false, nil, "cookie校验失败")
		return
	}
	// 获取用户信息
	ub, isExist, err := models.FindUserByIdentity(utils.DB, uc.UserId)
	if !isExist {
		utils.RespOK(writer, ApiModels.USERNOTEXIST, false, nil, "用户不存在")
		return
	}
	// 绑定请求参数
	var req ApiModels.CreateFileReqAPI
	err = c.ShouldBindJSON(&req)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	// 仅支持excel，word，ppt的创建
	if req.ExtendName != "xlsx" && req.ExtendName != "docx" && req.ExtendName != "pptx" {
		utils.RespOK(writer, ApiModels.FILETYPENOTSUPPORT, false, nil, "文件类型不支持")
		return
	}
	// 开启事务
	// todo:思考，事务A和B都进行在同一个路径创建同名文件，且在事务开启时都没找到同名文件，此时两个事务执行后是否会创建两个相同记录？
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		// 查询父文件夹记录
		parentDir, isExist, err := models.FindParentDirFromAbsPath(tx, ub.UserId, req.FilePath)
		if err != nil {
			utils.RespOK(writer, ApiModels.DATABASEERROR, false, nil, "创建文件夹失败")
			return errors.New("database error" + err.Error())
		}
		if !isExist {
			utils.RespOK(writer, ApiModels.PARENTNOTEXIST, false, nil, "无法找到父文件夹")
			return errors.New("parent directory not exist")
		}

		// 检查是否有重名文件
		_, isExist, err = models.FindFileByNameAndPath(tx, ub.UserId, req.FilePath, req.FileName, req.ExtendName)
		if err != nil {
			utils.RespOK(writer, ApiModels.FILEREPEAT, false, nil, "文件在当前文件夹已存在")
			return errors.New("file repeat")
		}
		if isExist {
			utils.RespOK(writer, ApiModels.FILEREPEAT, false, nil, "文件在当前文件夹已存在")
			return errors.New("file repeat")
		}
		// 创建文件
		userFileUUID := utils.GenerateUUID()
		poolFileUUID := utils.GenerateUUID()
		savePath := "./repository/upload_file/" + poolFileUUID
		file, err := os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			utils.RespOK(writer, ApiModels.FILECREATEERROR, false, nil, "文件保存出错")
			return err
		}
		err = file.Close()
		if err != nil {
			utils.RespOK(writer, ApiModels.FILESAVEERROR, false, nil, "文件保存出错")
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
		utils.RespOK(writer, 0, true, nil, "创建文件成功")
		return nil
	})

}

// CreateFolder 文件上传
func CreateFolder(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, isExist := utils.CheckCookie(c)
	if !isExist {
		utils.RespOK(writer, ApiModels.UNAUTHORIZED, false, nil, "cookie校验失败")
		return
	}
	var r ApiModels.CreateFolderRequest
	err := c.ShouldBind(&r)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	// 开启事务
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		// 获取用户信息
		ub, isExist, err := models.FindUserByIdentity(tx, uc.UserId)
		if !isExist {
			utils.RespOK(writer, ApiModels.USERNOTEXIST, false, nil, "用户不存在")
			return errors.New("user not exist")
		}
		// 查询父文件夹记录
		parentDir, isExist, err := models.FindParentDirFromAbsPath(tx, ub.UserId, r.FolderPath)
		if err != nil {
			utils.RespOK(writer, ApiModels.DATABASEERROR, false, nil, "创建文件夹失败")
			return errors.New("database error" + err.Error())
		}
		if !isExist {
			utils.RespOK(writer, ApiModels.PARENTNOTEXIST, false, nil, "无法找到父文件夹")
			return errors.New("parent directory not exist")
		}
		// 不需要查询文件是否存在，因为user_repository表中将(`user_id`,`parent_id`,`file_name`,`extend_name`,`file_type`)作为唯一索引
		// 因此，文件是否重复交由数据库是否返回错误代码是否为1062进行判断
		/*
			// 查询父文件夹下同名文件夹
			res := tx.Where("user_id = ? AND file_name = ? AND file_path = ? AND file_type = ?", ub.UserId, r.FolderName, r.FolderPath, utils.DIRECTORY).
				Find(&models.UserRepository{})
			if res.Error != nil {
				return res.Error
			}
			// 文件存在
			if res.RowsAffected != 0 {
				utils.RespOK(writer, ApiModels.FILEREPEAT, false, nil, "同名文件夹已存在")
				return errors.New("file repeat")
			}
		*/
		// 新增文件记录
		err = tx.Create(&models.UserRepository{
			UserFileId: utils.GenerateUUID(),
			UserId:     ub.UserId,
			FilePath:   r.FolderPath,
			ParentId:   parentDir.UserFileId,
			FileName:   r.FolderName,
			FileType:   utils.DIRECTORY,
			IsDir:      1,
			ExtendName: "",
			ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
			UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
		}).Error
		if err != nil {
			if utils.IsDuplicateEntryErr(err) {
				utils.RespOK(writer, ApiModels.FILEREPEAT, false, nil, "文件夹已存在")
				return err
			}
			utils.RespOK(writer, ApiModels.DATABASEERROR, false, nil, "创建文件夹失败")
			return err
		}
		utils.RespOK(writer, ApiModels.SUCCESS, true, nil, "创建文件夹成功")
		return nil
	})
}

func DeleteFile(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	ub, err := models.GetUserFromCoookie(utils.DB, c)
	if err != nil {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
	}

	type DeleteFileRequest struct {
		UserFileId string `json:"userFileId"`
	}
	var r DeleteFileRequest
	err = c.ShouldBind(&r)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	// 如果文件不存在，删除失败
	ur, isExist := models.FindUserFileById(ub.UserId, r.UserFileId)
	if !isExist {
		// 找不到记录
		utils.RespOK(writer, 1, false, nil, "文件不存在")
		return
	}
	// 开启事务，删除文件夹
	delBatchId := utils.GenerateUUID()
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		if ur.FileType == utils.DIRECTORY { // 如果文件是文件夹
			// 递归进入文件夹，删除文件夹内部的文件
			err = models.DelAllFilesFromDir(delBatchId, ub.UserId, ur.FilePath, ur.FileName)
			if err != nil {
				return err
			}
			// 删除文件夹自己
			err = utils.DB.Where("user_file_id = ?", ur.UserFileId).
				Delete(&models.UserRepository{}).Error
			if err != nil {
				return err
			}
		} else {
			err = utils.DB.Where("user_id = ? and user_file_id = ?", ub.UserId, r.UserFileId).
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
		utils.RespOK(writer, 0, false, nil, "删除文件失败")
		return
	}
	utils.RespOK(writer, 0, true, nil, "删除成功")
}

func DeleteFilesInBatch(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
	}
	// 获取用户信息
	ub, isExist, err := models.FindUserByIdentity(utils.DB, uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
	}

	type DeleteFilesRequest struct {
		UserFileIds string `json:"userFileIds"`
	}

	var r DeleteFilesRequest
	err = c.ShouldBind(&r)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	userFileIdList := strings.Split(r.UserFileIds, ",")
	// 开启事务，删除文件
	delBatchId := utils.GenerateUUID()
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		// 找出这些文件信息
		var urList []*models.UserRepository
		err = utils.DB.
			//Clauses(clause.Locking{Strength: "UPDATE"}). // 排他锁
			Where("user_id = ? and user_file_id in ?", ub.UserId, userFileIdList).
			Find(&urList).Error
		if err != nil {
			return err
		}
		// 循环文件
		for _, ur := range urList {
			if ur.FileType == utils.DIRECTORY {
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
		utils.RespOK(writer, 9999, false, nil, "删除文件失败")
		return
	}
	utils.RespOK(writer, 0, true, nil, "删除成功")
}

func FilePreview(c *gin.Context) {
	writer := c.Writer

	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	fmt.Fprintf(gin.DefaultWriter, "%v", uc)

	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}
	// 获取用户信息
	ub, isExist, err := models.FindUserByIdentity(utils.DB, uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
		return
	}

	// 处理请求参数
	//type FilePreviewRequest struct {
	//	UserFileId     string `json:"userFileId"`
	//	ShareBatchNum  string `json:"shareBatchNum"`
	//	ExtractionCode string `json:"extractionCode"`
	//	IsMin          bool   `json:"isMin"`
	//}
	//r := FilePreviewRequest{}
	//err := c.ShouldBindQuery(&r)
	//if err != nil {
	//	utils.RespBadReq(writer, "请求参数错误")
	//}
	userFileId := c.Query("userFileId")
	isMin := c.Query("isMin")

	// 获取文件信息
	ur, isExist1 := models.FindUserFileById(uc.UserId, userFileId)
	rp, isExist2 := models.FindRepFileByUserFileId(ub.UserId, userFileId)

	if !(isExist1 && isExist2) {
		utils.RespBadReq(writer, "文件不存在，请联系管理员")
		return
	}

	previewFilePath := rp.Path
	if isMin == "true" {
		// 预览最小文件
		switch ur.FileType {
		case utils.IMAGE:
			previewFilePath = rp.Path + "-pv"
		case utils.VIDEO:
			previewFilePath = rp.Path + "-pv"
		}
	}
	file, err := os.OpenFile(previewFilePath, os.O_RDONLY, 0777)
	defer file.Close()
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		utils.RespBadReq(writer, "出现错误1")
		return
	}
	writer.WriteHeader(http.StatusOK)
	return
}
