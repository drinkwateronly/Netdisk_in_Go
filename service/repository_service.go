package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"log"
	"net/http"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetUserStorage
// @Summary 获取用户存储容量
// @Produce json
// @Success 200 {object} string "存储容量"
// @Failure 400 {object} string "cookie校验失败"
// @Router /filetransfer/getstorage [get]
func GetUserStorage(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
	}
	// 获取用户信息
	ub, _ := models.FindUserByIdentity(uc.UserId)
	//ub, isExist := models.FindUserByIdentity(uc.UserId)
	//if !isExist {
	//	utils.RespBadReq(writer, "用户不存在")
	//}
	utils.RespOK(writer, 0, true, gin.H{
		"storageSize":      ub.StorageSize,
		"totalStorageSize": ub.TotalStorageSize,
	}, "存储容量")
}

// GetUserFileList 获取用户文件列表
func GetUserFileList(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
	}
	// 获取用户信息
	ub, isExist := models.FindUserByIdentity(uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
	}

	// 获取请求参数
	filePath := c.Query("filePath")
	fileType, err := strconv.Atoi(c.Query("fileType"))
	if err != nil {
		utils.RespBadReq(writer, "参数错误")
	}
	currentPage, err := strconv.Atoi(c.Query("currentPage"))
	if err != nil {
		utils.RespBadReq(writer, "参数错误")
		return
	}
	pageCount, err := strconv.Atoi(c.Query("pageCount"))
	if err != nil {
		utils.RespBadReq(writer, "参数错误")
		return
	}

	var files []models.UserRepository
	if fileType == 0 {
		files, err = models.FindFilesByPathAndPage(filePath, ub.UserId, currentPage, pageCount)
	} else {
		files, err = models.FindFilesByTypeAndPage(fileType, ub.UserId, currentPage, pageCount)
	}

	if err != gorm.ErrRecordNotFound && err != nil {
		utils.RespBadReq(writer, "参数错误")
		return
	}
	utils.RespOkWithDataList(writer, 0, files, len(files), "文件列表")
}

func FileUploadPrepare(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
	}
	// 获取用户信息
	ub, isExist := models.FindUserByIdentity(uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
	}

	// 文件分块参数
	//chunkNumber := c.Query("chunkNumber")
	//chunkNumber := c.Query("chunkNumber")
	//currentChunkSize := c.Query("currentChunkSize")
	//totalChunks := c.Query("totalChunks")

	// 文件信息
	totalSize, err := strconv.ParseInt(c.Query("totalSize"), 10, 64)
	if err != nil {
		utils.RespBadReq(writer, "参数错误")
		return
	}

	hash := c.Query("identifier")
	fileName := c.Query("filename")
	//relativePath := c.Query("relativePath")
	filePath := c.Query("filePath")
	isDir, err := strconv.Atoi(c.Query("isDir"))
	if err != nil {
		utils.RespBadReq(writer, "参数错误")
		return
	}
	// 判断存储空间是否足够，前端已经做好了此判断工作。
	if ub.StorageSize+totalSize > ub.TotalStorageSize {
		utils.RespBadReq(writer, "存储空间不足")
		return
	}
	// 处理出文件名、拓展名、文件类型
	split := strings.Split(fileName, ".")
	fileName = strings.Join(split[0:len(split)-1], ".")
	extendName := split[len(split)-1]
	fileType := utils.FileTypeId[extendName]
	if isDir == 1 {
		fileType = 6 // 文件夹
	} else if fileType == 0 {
		fileType = 5 // 其他
	}

	// 判断文件在当前文件夹是否重名
	if _, isExist := models.FindFileByNameAndPath(ub.UserId, filePath, fileName, extendName); isExist {
		utils.RespOK(writer, 999999, false, nil, "文件在当前文件夹已存在")
		return
	}

	// 如果文件大小为0，则上传文件
	if totalSize == 0 {
		utils.RespOK(writer, 0, true, gin.H{"skipUpload": false}, "开始上传文件")
		return
	}
	// 根据md5值判断文件在中心存储池是否已存在
	rp, isExist := models.FindFileByMD5(hash)
	if !isExist { // 文件不存在，上传文件
		utils.RespOK(writer, 0, true, gin.H{"skipUpload": false}, "开始上传文件")
		return
	}
	// 文件存在，进行秒传
	// 只有最后一个点后是文件拓展名filename.filename.ext
	ur := models.UserRepository{
		UserFileId: utils.GenerateUUID(),                     // 用户文件id
		UserId:     ub.UserId,                                // 用户id
		FileId:     rp.FileId,                                // 存储池文件id
		FilePath:   filePath,                                 // 文件存储路径
		FileName:   fileName,                                 // 文件名
		FileType:   fileType,                                 // 文件类型
		ExtendName: extendName,                               // 文件拓展名
		IsDir:      0,                                        // 是否是文件夹
		ModifyTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
		UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
		FileSize:   totalSize,                                // 文件尺寸
	}
	// 开启事务
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := utils.DB.Create(&ur).Error; err != nil {
			utils.RespBadReq(writer, "出现错误")
			return err
		}
		newStorageSize := ub.StorageSize + totalSize
		if err := tx.Model(&models.UserBasic{}).Select("storage_size").
			Where("user_id = ?", ub.UserId).
			Updates(models.UserBasic{StorageSize: newStorageSize}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
	}
	utils.RespOK(writer, 0, true, gin.H{"skipUpload": true}, "文件秒传")
}

// FileUpload 文件上传
func FileUpload(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
	}

	// 上传的文件参数
	chunkNumber, _ := strconv.Atoi(c.PostForm("chunkNumber")) // 当前分片的index
	//currentChunkSize := c.PostForm("currentChunkSize")  // 当前分片的大小，未使用
	totalSize, _ := strconv.ParseInt(c.PostForm("totalSize"), 10, 64)
	totalChunks, _ := strconv.Atoi(c.PostForm("totalChunks"))

	fileMD5 := c.PostForm("identifier")        // 文件哈希值
	rqfileName := c.PostForm("filename")       // 文件名，包括拓展名
	filePath := c.PostForm("filePath")         // 文件在用户存储区的
	isDir := c.PostForm("isDir")               // 是否是文件夹
	relativePath := c.PostForm("relativePath") // 相对路径，暂未使用

	chunkPath := fmt.Sprintf("./repository/chunk_file/%s-%d.chunk", fileMD5, chunkNumber)
	if utils.IsFileExist(chunkPath) {
		utils.RespOK(writer, 0, true, nil, "分片上传成功")
		return
	}

	// 保存分块文件
	uploadedFile, err := c.FormFile("file")
	if uploadedFile == nil {
		utils.RespOK(writer, 99999, false, nil, "出错或用户取消上传")
		return
	}
	err = c.SaveUploadedFile(uploadedFile, chunkPath)
	if err != nil {
		log.Println(err)
		utils.RespBadReq(writer, "出现错误")
		return
	}

	if chunkNumber != totalChunks {
		utils.RespOK(writer, 0, true, nil, "分片上传成功")
		return
	}

	// 走到这里意味着最后一块分块上传完成，开始合并文件

	// 获取文件基础信息
	split := strings.Split(rqfileName, ".")
	fileName := strings.Join(split[0:len(split)-1], ".")
	// 只有最后一个点后是文件拓展名filename.filename.ext
	extendName := split[len(split)-1]
	fileType := utils.FileTypeId[extendName]
	if isDir == "1" {
		fileType = utils.DIRECTORY // 文件夹
	} else if fileType == 0 {
		fileType = utils.OTHER // 其他
	}

	// 获取用户信息
	ub, isExist := models.FindUserByIdentity(uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
	}

	// 此时文件以相对路径形式上传，这种形式常见于整个文件夹的上传
	// 例如123/456/OnlyOffice.vue，接下来的步骤将按顺序创建文件夹123和456
	if relativePath != rqfileName {
		// 取出 [123, 456]，即文件相对路径先后进入的文件夹的列表
		folderList := strings.Split(relativePath[:len(relativePath)-len(rqfileName)-1], "/")
		// 若文件夹不存在，则创建，若存在，则继续
		fmt.Fprintln(gin.DefaultWriter, "not exist file", relativePath)
		for _, folderName := range folderList {
			// 开启事务
			err = utils.DB.Transaction(func(tx *gorm.DB) error {
				// 当前文件上传的目录filePath有没有名为folderName的文件夹
				res := utils.DB.Clauses( // 加入排他锁
					clause.Locking{
						Strength: "UPDATE",
					},
				).
					Where("user_id = ? AND file_name = ? AND file_path = ? AND file_type = ?", ub.UserId, folderName, filePath, utils.DIRECTORY).
					Find(&models.UserRepository{})
				// 文件夹不存在，就创建在路径filePath的文件夹folderName
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 {
					err = utils.DB.
						//Set("gorm: query_option", "FOR UPDATE").
						Create(&models.UserRepository{
							UserFileId: utils.GenerateUUID(),
							UserId:     ub.UserId,
							FilePath:   filePath,
							FileName:   folderName,
							FileType:   utils.DIRECTORY,
							IsDir:      1,
							ExtendName: "",
							ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
							UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
						}).Error
					if err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				utils.RespBadReq(writer, "创建文件夹出错")
				return
			}

			// 然后进入下一级目录，继续创建文件夹
			if filePath == "/" {
				filePath += folderName
			} else {
				filePath += "/" + folderName
			}
		}
	}
	// 最后filePath变成所要上传的文件的绝对路径，上例中，则为/123/456/

	// 生成文件uuid
	poolFileId := utils.GenerateUUID()
	userFileId := utils.GenerateUUID()
	savePath := "./repository/upload_file/" + poolFileId

	// 将分块文件合并
	err = utils.MergeChunksToFile(fileMD5, poolFileId, totalChunks)
	if err != nil {
		utils.RespOK(writer, 99999, false, nil, "融合分片文件时出错")
		return
	}

	// 校验md5
	mergeMD5, err := utils.GetFileMD5FromPath(savePath)
	if mergeMD5 != fileMD5 || err != nil {
		utils.RespOK(writer, 99999, false, nil, "md5校验出错，文件上传失败")
		return
	}

	// 开始写入数据库
	ur := models.UserRepository{
		UserFileId: userFileId, // 用户文件id
		UserId:     ub.UserId,  // 用户id
		FileId:     poolFileId, // 存储池文件id
		FilePath:   filePath,   //
		FileName:   fileName,   // 用户存储时的文件名
		FileType:   fileType,   // 文件类型
		ExtendName: extendName, // 文件拓展名
		FileSize:   totalSize,  // 文件大小
		IsDir:      0,
		ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
		UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
	}
	rp := models.RepositoryPool{
		FileId: poolFileId,
		Hash:   fileMD5, // todo:mergeFileMD5
		Size:   totalSize,
		Path:   savePath,
	}
	// 开启事务，插入文件记录repository_pool, user_repository，修改用户存储容量
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ur).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		if err := tx.Create(&rp).Error; err != nil {
			return err
		}
		newStorageSize := ub.StorageSize + totalSize
		if err := tx.Model(&models.UserBasic{}).Select("storage_size").
			Where("user_id = ?", ub.UserId).
			Updates(models.UserBasic{StorageSize: newStorageSize}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	utils.RespOK(writer, 0, true, nil, "文件上传成功")
	// resp后，用户已经收到文件上传的结果
	// 如果文件类型是图片/视频，则保存preview格式，方便后续前端预览
	switch ur.FileType {
	case utils.IMAGE:
		// 不处理错误
		_ = utils.SavePreviewFromImage(savePath, extendName)
	case utils.VIDEO:
		// 不处理错误
		_ = utils.SavePreviewFromVideo(savePath, 5)
	}
	// 开始删除分片文件
	utils.DeleteAllChunks(fileMD5, totalChunks)
}

func CreateFolder(c *gin.Context) {
	type CreateFolderRequest struct {
		FolderName string `json:"fileName"`
		FolderPath string `json:"filePath"`
	}

	writer := c.Writer
	// 校验cookie
	uc, _ := utils.CheckCookie(c)
	// 获取用户信息
	ub, isExist := models.FindUserByIdentity(uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
		return
	}

	var r CreateFolderRequest
	err := c.ShouldBind(&r)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	// 检查是否有重名文件
	rowsAffected := utils.DB.
		Where("user_id = ? AND file_name = ? AND file_path = ? AND file_type = ?", ub.UserId, r.FolderName, r.FolderPath, utils.DIRECTORY).
		Find(&models.UserRepository{}).
		RowsAffected
	if rowsAffected != 0 {
		utils.RespOK(writer, 999999, false, nil, "同名文件夹已存在")
		return
	}

	err = utils.DB.Create(&models.UserRepository{
		UserFileId: utils.GenerateUUID(),
		UserId:     ub.UserId,
		FilePath:   r.FolderPath,
		FileName:   r.FolderName,
		FileType:   utils.DIRECTORY,
		IsDir:      1,
		ExtendName: "",
		ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
		UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
	}).Error
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	utils.RespOK(writer, 0, true, nil, "创建文件夹成功")
}

// CreateFile 创建文件
func CreateFile(c *gin.Context) {
	type CreateFileRequest struct {
		FileName   string `json:"fileName"`
		FilePath   string `json:"filePath"`
		ExtendName string `json:"extendName"`
	}
	writer := c.Writer
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}
	// 获取用户信息
	ub, isExist := models.FindUserByIdentity(uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
	}
	// 获取参数
	var r CreateFileRequest
	err := c.ShouldBind(&r)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	// 检查是否有重名文件
	if _, isExist := models.FindFileByNameAndPath(ub.UserId, r.FilePath, r.FileName, r.ExtendName); isExist {
		utils.RespOK(writer, 999999, false, nil, "文件在当前文件夹已存在")
		return
	}
	// 创建文件
	userFileUUID := utils.GenerateUUID()
	poolFileUUID := utils.GenerateUUID()
	savePath := "./repository/upload_file/" + poolFileUUID
	file, err := os.OpenFile(savePath, os.O_CREATE, 0777)
	file.Close()

	ur := &models.UserRepository{
		UserFileId: userFileUUID,
		UserId:     ub.UserId,
		FileId:     poolFileUUID,
		FilePath:   r.FilePath,
		FileName:   r.FileName,
		FileType:   2,
		IsDir:      0,
		ExtendName: r.ExtendName,
		ModifyTime: time.Now().Format("2006-01-02 15:04:05"),
		UploadTime: time.Now().Format("2006-01-02 15:04:05"), // 上传时间
		FileSize:   0,
	}
	rp := models.RepositoryPool{
		FileId: poolFileUUID,
		Hash:   "d41d8cd98f00b204e9800998ecf8427e", // todo:mergeFileMD5
		Size:   0,
		Path:   savePath,
	}
	// 开启事务，插入文件记录repository_pool, user_repository，修改用户存储容量
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ur).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		if err := tx.Create(&rp).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}

	utils.RespOK(writer, 0, true, nil, "创建文件成功")

}

func DeleteFile(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	ub, err := models.GetUserFromCoookie(c)
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
	ur, isExist := models.FindFileById(ub.UserId, r.UserFileId)
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
	ub, isExist := models.FindUserByIdentity(uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
	}

	type DeleteFilesRequest struct {
		UserFileIds string `json:"userFileIds"`
	}

	var r DeleteFilesRequest
	err := c.ShouldBind(&r)
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
			Clauses(clause.Locking{Strength: "UPDATE"}). // 排他锁
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

func FileDownload(c *gin.Context) {
	writer := c.Writer

	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
	}
	// 获取用户信息
	ub, isExist := models.FindUserByIdentity(uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
	}
	// 处理请求参数
	//type FileDownloadRequest struct {
	//	UserFileId     string `json:"userFileId"`
	//	ShareBatchNum  string `json:"shareBatchNum"`
	//	ExtractionCode string `json:"extractionCode"`
	//}
	//r := FileDownloadRequest{}
	//err := c.ShouldBindQuery(&r)
	//if err != nil {
	//	utils.RespFail(writer, http.StatusBadRequest, "请求参数错误")
	//}

	userFileId := c.Query("userFileId")
	//userFileId = c.PostForm("userFileId")
	//shareBatchNum:= c.Query("userFileId")
	//extractionCode := c.Query("userFileId")

	// 获取文件
	rp, isExist := models.FindFileSavePathById(ub.UserId, userFileId)
	if !isExist {
		utils.RespBadReq(writer, "文件不存在")
		return
	}

	file, err := os.OpenFile(rp.Path, os.O_RDONLY, 0777)
	defer file.Close()
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	utils.RespOK(writer, 0, true, nil, "下载成功")

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
	ub, isExist := models.FindUserByIdentity(uc.UserId)
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
	ur, isExist1 := models.FindFileById(uc.UserId, userFileId)
	rp, isExist2 := models.FindFileSavePathById(ub.UserId, userFileId)

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

// 查看回收站文件
func GetRecoveryList(c *gin.Context) {
	writer := c.Writer
	// 校验cookie，获取用户信息
	ub, err := models.GetUserFromCoookie(c)
	if err != nil {
		utils.RespBadReq(writer, "用户校验失败")
		return
	}
	var recoveryFiles []models.RecoveryBatch
	utils.DB.
		Where("user_id = ?", ub.UserId).
		Find(&recoveryFiles)
	utils.RespOkWithDataList(writer, 0, recoveryFiles, len(recoveryFiles), "文件列表")
}
