package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"log"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetUserStorage(c *gin.Context) {
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
	if _, isExist := models.FindFileByPathAndName(filePath, fileName, extendName, ub.UserId); isExist {
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
		UserFileId: utils.GenerateUUID(), // 用户文件id
		UserId:     ub.UserId,            // 用户id
		FileId:     rp.FileId,            // 存储池文件id
		IsDir:      isDir,                // 是否是目录
		FilePath:   filePath,             // 文件存储路径
		FileName:   fileName,             // 文件名
		FileType:   fileType,             // 文件类型
		ExtendName: extendName,           // 文件拓展名
		UploadTime: time.Now(),           // 上传时间
		FileSize:   totalSize,            // 文件尺寸
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
	// 获取用户信息
	ub, isExist := models.FindUserByIdentity(uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
	}

	// 上传的文件参数
	chunkNumber, _ := strconv.Atoi(c.PostForm("chunkNumber")) // 当前分片的index
	//currentChunkSize := c.PostForm("currentChunkSize")  // 当前分片的大小，未使用
	totalSize, _ := strconv.ParseInt(c.PostForm("totalSize"), 10, 64)
	totalChunks, _ := strconv.Atoi(c.PostForm("totalChunks"))

	fileMD5 := c.PostForm("identifier") // 文件哈希值
	fileName := c.PostForm("filename")  // 文件名，包括拓展名
	//relativePath := c.PostForm("relativePath")    // 保存的相对路径
	filePath := c.PostForm("filePath")            // 文件在用户存储区的
	isDir, _ := strconv.Atoi(c.PostForm("isDir")) // 是否是文件夹

	// 只有最后一个点后是文件拓展名filename.filename.ext
	split := strings.Split(fileName, ".")
	fileName = strings.Join(split[0:len(split)-1], ".")
	extendName := split[len(split)-1]
	fileType := utils.FileTypeId[extendName]
	if isDir == 1 {
		fileType = utils.DICTIONARY // 文件夹
	} else if fileType == 0 {
		fileType = utils.OTHER // 其他
	}

	// 保存分块文件
	uploadedFile, err := c.FormFile("file")
	err = c.SaveUploadedFile(uploadedFile, fmt.Sprintf("./repository/chunk_file/%s-%d.chunk", fileMD5, chunkNumber))
	if err != nil {
		log.Println(err)
		utils.RespBadReq(writer, "出现错误")
		return
	}

	if chunkNumber != totalChunks {
		utils.RespOK(writer, 0, true, nil, "分块上传成功")
		return
	}

	// 走到这里意味着最后一块分块上传完成
	poolFileId := utils.GenerateUUID()
	userFileId := utils.GenerateUUID()
	savePath := fmt.Sprintf("./repository/upload_file/%s", poolFileId)
	// 将分块文件合并
	mergeFileMD5, err := utils.MergeChunkToFile(fileMD5, poolFileId, totalChunks)
	_ = mergeFileMD5
	//todo:实际上是需要对比两个md5值，判断文件是否上传成功，
	//todo:但前端使用的spark-md5和后端的crypto包md5计算出来的值不同，暂时没找到解决方案。
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	// 开始写入数据库
	ur := models.UserRepository{
		UserFileId: userFileId, // 用户文件id
		UserId:     ub.UserId,  // 用户id
		FileId:     poolFileId, // 存储池文件id
		IsDir:      isDir,      // 是否是目录
		FilePath:   filePath,   //
		FileName:   fileName,   // 用户存储时的文件名
		FileType:   fileType,   // 文件类型
		ExtendName: extendName, // 文件拓展名
		FileSize:   totalSize,  // 文件大小
		UploadTime: time.Now(),
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
	}

	err = utils.DB.Create(&models.UserRepository{
		UserFileId: utils.GenerateUUID(),
		UserId:     ub.UserId,
		IsDir:      1,
		FilePath:   r.FolderPath,
		FileName:   r.FolderName,
		FileType:   6,
		ExtendName: "",
		UploadTime: time.Now(),
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

	// 创建文件
	userFileUUID := utils.GenerateUUID()
	file, err := os.OpenFile("./repository/upload_file/"+userFileUUID, os.O_CREATE, 0777)
	file.Close()
	// 由于新建的文件的size为0，所以算出来的hash都一样，没必要放到中心存储池。
	err = utils.DB.Create(&models.UserRepository{
		UserFileId: userFileUUID,
		UserId:     ub.UserId,
		FileId:     "",
		IsDir:      0,
		FilePath:   r.FilePath,
		FileName:   r.FileName,
		FileType:   2,
		ExtendName: r.ExtendName,
		UploadTime: time.Now(),
		FileSize:   0,
	}).Error
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	utils.RespOK(writer, 0, true, nil, "创建文件成功")
}

func DeleteFile(c *gin.Context) {
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

	type DeleteFileRequest struct {
		UserFileId string `json:"userFileId"`
	}
	var r DeleteFileRequest
	err := c.ShouldBind(&r)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	err = utils.DB.Where("user_id = ? and user_file_id = ?", ub.UserId, r.UserFileId).
		Delete(&models.UserRepository{}).Error
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
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
	UserFileIdList := strings.Split(r.UserFileIds, ",")

	// db.Delete(&users, []int{1,2,3}) DELETE FROM users WHERE id IN (1,2,3);
	err = utils.DB.Where("user_id = ? and user_file_id in ?", ub.UserId, UserFileIdList).
		Delete(&models.UserRepository{}).Error
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
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
	type FilePreviewRequest struct {
		UserFileId     string `json:"userFileId"`
		ShareBatchNum  string `json:"shareBatchNum"`
		ExtractionCode string `json:"extractionCode"`
		IsMin          bool   `json:"isMin"`
	}
	r := FilePreviewRequest{}
	err := c.ShouldBindQuery(&r)
	if err != nil {
		utils.RespBadReq(writer, "请求参数错误")
	}

	//json := make(map[string]interface{})
	//c.BindJSON(&json)
	//userFileId := json["userFileId"].(string)
	userFileId := c.Query("userFileId")
	isMin := c.Query("isMin")
	//extractionCode := c.Query("userFileId")

	// 预览最小文件
	if isMin == "true" {
		utils.RespOK(writer, 0, true, nil, "最小化预览")
		return
	}

	// 预览整体文件
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
	utils.RespOK(writer, 0, true, nil, "原始预览")
	return

}
