package file_service

import (
	"errors"
	"fmt"
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

// FileUploadPrepare
// @Summary 文件上传预备
// @Produce json
// @Param req body api.FileUploadReqAPI true "文件上传请求"
// @Success 200 {object} string "存储容量"
// @Failure 400 {object} string "参数出错"
// @Router /filetransfer/uploadfileprepare [GET]
func FileUploadPrepare(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)

	// 绑定query请求参数
	var req api.FileUploadReqAPI
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数出错")
		return
	}
	// 判断存储空间是否足够
	if ub.StorageSize+req.TotalSize > ub.TotalStorageSize {
		response.RespOKFail(writer, response.StorageNotEnough, "用户存储空间不足")
		return
	}
	// 如果文件大小为0，则上传文件
	if req.TotalSize == 0 {
		response.RespOK(writer, 0, true, gin.H{"skipUpload": false}, "开始上传文件")
		return
	}
	// 根据md5值判断文件在中心存储池是否已存在
	rp, isExist := models.FindFileByMD5(req.FileMD5)
	if !isExist { // 文件不存在，上传文件
		response.RespOK(writer, 0, true, gin.H{"skipUpload": false}, "开始上传文件")
		return
	}

	// 执行至此表示中心存储池文件存在，应当进行文件秒传

	// 处理出文件名、拓展名、文件的逻辑绝对路径、文件类型
	fileInfo := filehandler.GetFileInfoFromReq(req)
	var parentId string     // 记录当前文件父文件夹id
	curPath := req.FilePath // 当前路径就是文件上传时候的根路径

	// 进行秒传，即开启事务为user_repository添加记录即可
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 查存储文件的上级文件夹，以为当前秒传文件获取parent_id
		parentDir, err := models.FindParentDirFromFilePath(tx, ub.UserId, fileInfo.FilePath)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			// 数据库出错
			response.RespOKFail(writer, response.DatabaseError, err.Error())
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 存储文件的上级文件夹不存在，这种情况常见于对整个文件夹的上传
			// 例如：在/123为根目录上传456/789/0.txt，而/123目录下没有456文件夹，自然也没有789文件夹。
			// 接下来的步骤将在文件夹123按顺序创建文件夹456和789
			uploadRoot, err := models.FindParentDirFromFilePath(tx, ub.UserId, curPath) // 找到/123的文件id
			if err != nil {
				// gorm.ErrRecordNotFound 或其他数据库错误
				response.RespOKFail(writer, response.DatabaseError, err.Error())
				return err
			}
			// 获得上传时的根目录的用户文件id，即/123的用户文件id
			parentId = uploadRoot.UserFileId

			// 得到相对路径"456/789"
			relativePath := req.RelativePath[:len(req.RelativePath)-len(req.FileName)-1]

			// 取出文件夹列表 [456, 789]，即文件相对路径先后进入的文件夹的列表
			folderList := strings.Split(relativePath, "/")

			// 接下来for循环中，进入curPath的文件夹，查询有没有folderName的文件夹，有则修改curPath进入下一级文件夹，无则创建文件夹folderName。
			for _, folderName := range folderList {
				var folder models.UserRepository
				// 当前文件上传的目录filePath有没有名为folderName的文件夹
				res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("user_id = ? AND file_name = ?  AND is_dir = 1 AND file_path = ?", ub.UserId, folderName, curPath).
					Find(&folder)
				// 文件夹不存在，就创建在路径filePath的文件夹folderName
				if res.RowsAffected == 0 {
					folder = models.UserRepository{
						UserFileId: common.GenerateUUID(),
						UserId:     ub.UserId,
						FilePath:   curPath,
						FileName:   folderName,
						ParentId:   parentId,
						FileType:   filehandler.DIRECTORY,
						IsDir:      1,
						ExtendName: "",
						ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
						UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
					}
					err = tx.Create(&folder).Error
					//err = tx.Where("NOT EXIST ?").Create(&folder).Error
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
			// 存储文件的上级文件夹存在
			parentId = parentDir.UserFileId
		}

		// 此时存储文件的上级文件夹已存在，开始创建文件记录
		ur := models.UserRepository{
			UserFileId: common.GenerateUUID(),                    // 用户文件id
			UserId:     ub.UserId,                                // 用户id
			FileId:     rp.FileId,                                // 存储池文件id
			FilePath:   fileInfo.FilePath,                        //
			FileName:   fileInfo.FileName,                        // 用户存储时的文件名
			ExtendName: fileInfo.ExtendName,                      // 文件拓展名
			FileType:   fileInfo.FileType,                        // 文件拓展名
			ParentId:   parentId,                                 // 父文件夹的id
			IsDir:      0,                                        // 是否是文件夹
			ModifyTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
			UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
			FileSize:   req.TotalSize,                            // 文件尺寸
		}
		res := tx.Where("parent_id = ? AND file_name = ? AND extend_name = ? AND is_dir = ?",
			ur.ParentId, ur.FileName, ur.ExtendName, ur.IsDir).Find(&models.UserRepository{}) // 使用了联合索引
		if res.Error != nil {
			response.RespOK(writer, 9999, false, nil, "文件上传失败："+err.Error())
			return err
		}
		for res.RowsAffected != 0 {
			ur.FileName = filehandler.RenameConflictFile(ur.FileName)
			res = tx.Where("parent_id = ? AND file_name = ? AND extend_name = ? AND is_dir = ?",
				ur.ParentId, ur.FileName, ur.ExtendName, ur.IsDir). // 使用了联合索引
				Find(&models.UserRepository{})
			if res.Error != nil {
				response.RespOK(writer, 9999, false, nil, "文件上传失败："+err.Error())
				return err
			}
		}

		if err := tx.Create(&ur).Error; err != nil {
			return err
		}
		newStorageSize := ub.StorageSize + req.TotalSize
		if err := tx.Where("user_id = ?", ub.UserId).
			Updates(models.UserBasic{StorageSize: newStorageSize}).Error; err != nil {
			return err
		}
		response.RespOK(writer, 0, true, gin.H{"skipUpload": true}, "文件秒传")
		return nil
	})
	if err != nil {
		response.RespOK(writer, 9999, false, nil, "文件上传失败："+err.Error())
	}
}

// FileUpload 文件上传
// FileUpload
// @Summary 文件上传
// @Produce json
// @Param req body api.FileUploadReqAPI true "文件上传请求"
// @Router /filetransfer/uploadfile [POST]
func FileUpload(c *gin.Context) {
	var savePath string
	var ur models.UserRepository
	var err error
	writer := c.Writer

	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)

	// 绑定请求参数
	var req api.FileUploadReqAPI
	err = c.ShouldBind(&req)
	if err != nil {
		response.RespOK(writer, 999999, false, nil, "参数出错")
		return
	}

	// 判断存储空间是否足够
	if ub.StorageSize+req.TotalSize > ub.TotalStorageSize {
		response.RespOKFail(writer, response.StorageNotEnough, "用户存储空间不足")
		return
	}

	// 本次上传的文件分片保存位置
	chunkPath := fmt.Sprintf("./repository/chunk_file/%s-%d.chunk", req.FileMD5, req.ChunkNumber)
	/*
		//该分片可能之前上传过，例如上次上传时失败，此时略过上传
		//if !common.IsChunkExist(chunkPath, req.CurrentChunkSize) {
	*/

	// 获取分片
	uploadedFile, err := c.FormFile("file")
	if err != nil || uploadedFile == nil {
		response.RespOK(writer, 999999, false, nil, "上传出错")
		return
	}
	// 将分片文件保存到chunkPath中
	if err = c.SaveUploadedFile(uploadedFile, chunkPath); err != nil {
		response.RespOKFail(writer, 999999, "上传出错")
		return
	}
	// 如果不是最后一个分片，那么继续上传
	if req.ChunkNumber != req.TotalChunks {
		response.RespOKSuccess(writer, response.Success, nil, "分片上传成功")
		return
	}

	// #############走到这里意味着最后一块分块上传完成，开始合并文件################
	// 处理出文件名、拓展名、文件的逻辑绝对路径、文件类型
	fileInfo := filehandler.GetFileInfoFromReq(req)

	// 所有文件分片上传完成，并得到了文件信息，开启事务
	err = models.DB.Transaction(func(tx *gorm.DB) error {

		// 判断文件在当前文件夹是否重名
		//if _, isExist, _ := models.FindFileByNameAndPath(tx, ub.UserId,
		//	fileInfo.FilePath,
		//	fileInfo.FileName[:len(fileInfo.FileName)-len(fileInfo.ExtendName)-1],
		//	fileInfo.ExtendName); isExist {
		//	return errors.New("文件在当前文件夹已存在")
		//}
		//if err != nil {
		//	return errors.New("文件夹不存在")
		//}
		// 根据绝对路径 判断文件的父文件夹是否存在
		parentDir, err := models.FindParentDirFromFilePath(tx, ub.UserId, fileInfo.FilePath)
		var parentId string     // 记录当前文件/文件夹的父文件夹id
		curPath := req.FilePath // 当前路径就是文件上传时候的根路径
		// if成立时，存放上传文件的文件夹不存在，这种情况常见于整个文件夹的上传时存在相对路径
		// 例在/123目录上传456/789/0.txt，接下来的步骤将在文件夹123按顺序创建文件夹456和789
		// isExist原来是
		if err != nil {
			// 找到/123的文件id
			uploadRoot, err := models.FindParentDirFromFilePath(tx, ub.UserId, curPath)
			if err != nil {
				return err
			}
			parentId = uploadRoot.UserFileId

			// 得到相对路径"456/789"
			//strings.TrimSuffix(req.RelativePath, "/"+req.FileName)
			var relativePath string

			relativePathLen := len(req.RelativePath) - len(req.FileName)
			relativePath = req.RelativePath[:relativePathLen-1]

			// 取出文件夹列表 [456, 789]，即文件相对路径先后进入的文件夹的列表
			folderList := strings.Split(relativePath, "/")
			// 接下来for循环中，进入curPath的文件夹，查询有没有folderName的文件夹，有则修改curPath进入下一级文件夹，无则创建文件夹folderName。

			for _, folderName := range folderList {

				var folder models.UserRepository
				// 当前文件上传的目录filePath有没有名为folderName的文件夹
				res := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					Where("user_id = ? AND file_name = ?  AND is_dir = 1 AND file_path = ?", ub.UserId, folderName, curPath).
					Find(&folder)
				// 文件夹不存在，就创建在路径filePath的文件夹folderName
				if res.RowsAffected == 0 {
					folder = models.UserRepository{
						UserFileId: common.GenerateUUID(),
						UserId:     ub.UserId,
						FilePath:   curPath,
						FileName:   folderName,
						ParentId:   parentId,
						FileType:   filehandler.DIRECTORY,
						IsDir:      1,
						ExtendName: "",
						ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
						UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
					}
					err = tx.Create(&folder).Error
					//err = tx.Where("NOT EXIST ?").Create(&folder).Error
					if err != nil {
						return err
					}
				}

				// 写法2
				//folder := models.UserRepository{
				//	UserFileId: common.GenerateUUID(),
				//	UserId:     ub.UserId,
				//	FilePath:   curPath,
				//	FileName:   folderName,
				//	ParentId:   parentId,
				//	FileType:   filehandler.DIRECTORY,
				//	IsDir:      1,
				//	ExtendName: "",
				//	ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
				//	UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
				//}
				//// 如果文件夹不存在，就创建在路径filePath的文件夹folderName
				//err = tx.
				//	Where("user_id = ? AND file_name = ?  AND is_dir = 1 AND file_path = ?", ub.UserId, folderName, curPath).
				//	Clauses(clause.Locking{Strength: "UPDATE"}).
				//	FirstOrCreate(&folder).Error
				//if err != nil {
				//	return err
				//}

				parentId = folder.UserFileId // 新创建的文件夹的id作为下一级文件的parentId

				if curPath == "/" { // 然后进入下一级目录，继续创建文件夹
					curPath += folderName
				} else {
					curPath += "/" + folderName
				}
			}
		} else {
			// 文件的父文件夹存在，
			parentId = parentDir.UserFileId
		}
		// 最后filePath变成所要上传的文件的路径，上例中，则为/123/456/

		// 走到此处，表示存放上传文件的文件夹存在

		// 生成上传文件的uuid
		poolFileId := common.GenerateUUID()
		userFileId := common.GenerateUUID()
		savePath = "./repository/upload_file/" + poolFileId

		// 将分块文件合并
		err = filehandler.MergeChunksToFile(req.FileMD5, poolFileId, req.TotalChunks)
		if err != nil {
			return err
		}

		// 校验md5
		mergeMD5, err := common.GetFileMD5FromPath(savePath)
		if mergeMD5 != req.FileMD5 || err != nil {
			return err
		}

		// 开始写入数据库
		ur = models.UserRepository{
			UserFileId: userFileId, // 用户文件id
			FileId:     poolFileId, // 存储池文件id
			UserId:     ub.UserId,  // 用户id
			ParentId:   parentId,
			FilePath:   fileInfo.FilePath,   //
			FileName:   fileInfo.FileName,   // 用户存储时的文件名
			ExtendName: fileInfo.ExtendName, // 文件拓展名
			FileType:   fileInfo.FileType,   // 文件类型
			IsDir:      0,
			FileSize:   req.TotalSize, // 文件大小
			ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
			UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
		}
		res := tx.Where("parent_id = ? AND file_name = ? AND extend_name = ? AND is_dir = ?",
			ur.ParentId, ur.FileName, ur.ExtendName, ur.IsDir).Find(&models.UserRepository{}) // 使用了联合索引
		if res.Error != nil {
			response.RespOK(writer, 9999, false, nil, "文件上传失败："+err.Error())
			return err
		}
		for res.RowsAffected != 0 {
			ur.FileName = filehandler.RenameConflictFile(ur.FileName)
			res = tx.Where("parent_id = ? AND file_name = ? AND extend_name = ? AND is_dir = ?",
				ur.ParentId, ur.FileName, ur.ExtendName, ur.IsDir). // 使用了联合索引
				Find(&models.UserRepository{})
			if res.Error != nil {
				response.RespOK(writer, 9999, false, nil, "文件上传失败："+err.Error())
				return err
			}
		}
		// 插入文件记录repository_pool, user_repository，修改用户存储容量
		if err := tx.Create(&ur).Error; err != nil {
			return err
		}

		rp := models.RepositoryPool{
			FileId: poolFileId,
			Hash:   req.FileMD5,
			Size:   req.TotalSize,
			Path:   savePath,
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
		response.RespOK(writer, 9999, false, nil, "文件上传失败："+err.Error())
		return
	}

	response.RespOK(writer, 0, true, nil, "文件上传成功")
	// resp后，用户已经收到文件上传的结果
	// 如果文件类型是图片/视频，则保存preview格式，方便后续前端预览
	switch ur.FileType {
	case filehandler.IMAGE:
		// 不处理错误
		_ = filehandler.SavePreviewFromImage(savePath, fileInfo.ExtendName)
	case filehandler.VIDEO:
		// 不处理错误
		_ = filehandler.SavePreviewFromVideo(savePath, 5)
	}
	// 开始删除分片文件，不删除会好点，后续可以统一删除
	//common.DeleteAllChunks(fileMD5, totalChunks)
}
