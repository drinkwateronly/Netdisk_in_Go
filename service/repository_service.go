package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
	"strconv"
	"strings"
	"time"
)

func GetUserStorage(c *gin.Context) {
	writer := c.Writer
	token, err := c.Cookie("token")
	if err != nil {
		utils.RespFail(writer, "验证失败")
	}

	uc, err := utils.ParseToken(token)
	if err != nil {
		utils.RespFail(writer, "验证失败")
	}
	ub, isExist := models.FindUserByPhone(uc.Phone)
	if !isExist {
		utils.RespFail(writer, "验证失败")
	}
	utils.RespOK(writer, struct {
		StorageSize      int64 `json:"storageSize"`      // 已使用的存储容量
		TotalStorageSize int64 `json:"totalStorageSize"` // 总存储容量
	}{
		StorageSize:      ub.StorageSize,
		TotalStorageSize: ub.TotalStorageSize,
	}, "存储容量")
}

func GetUserFileList(c *gin.Context) {
	// 校验token
	writer := c.Writer
	uc, err := utils.ParseTokenFromCookie(c)
	if err != nil {
		utils.RespFail(writer, "验证失败")
		return
	}

	// 获取请求参数
	filePath := c.Query("filePath")
	fileType, err := strconv.Atoi(c.Query("fileType"))
	if err != nil {
		utils.RespFail(writer, "参数错误")
	}
	currentPage, err := strconv.Atoi(c.Query("currentPage"))
	if err != nil {
		utils.RespFail(writer, "参数错误")
	}
	pageCount, err := strconv.Atoi(c.Query("pageCount"))
	if err != nil {
		utils.RespFail(writer, "参数错误")
	}

	//
	files, err := models.FindFilesByPath(filePath, uc.Identity, fileType, currentPage, pageCount)
	if err != nil {
		utils.RespFail(writer, "验证失败")
	}
	utils.RespList(writer, files, len(files), "文件列表")
}

func PrepareFileUpload(c *gin.Context) {
	writer := c.Writer
	uc, err := utils.ParseTokenFromCookie(c)
	if err != nil {
		utils.RespFail(writer, "验证失败")
		return
	}
	ub, isExist := models.FindUserByPhone(uc.Phone)
	if !isExist {
		utils.RespFail(writer, "用户不存在")
	}
	// 获取要上传的文件参数
	//chunkNumber := c.Query("chunkNumber")
	//chunkNumber := c.Query("chunkNumber")
	//currentChunkSize := c.Query("currentChunkSize")
	totalSize, _ := strconv.ParseInt(c.Query("totalSize"), 10, 64)
	//identifier := c.Query("identifier")
	//filename := c.Query("filename")
	//relativePath := c.Query("relativePath")
	//totalChunks := c.Query("totalChunks")
	//filePath := c.Query("filePath")
	//isDir := c.Query("isDir")
	// 判断存储空间是否足够
	if ub.StorageSize+totalSize > ub.TotalStorageSize {
		utils.RespFail(writer, "内存空间不足")
	}
	// 判断文件在用户存储池是否已存在

	// 判断文件hash是否在中心存储池已存在
	utils.RespOK(writer, gin.H{
		"skipUpload": false,
	}, "1")
}

func FileUpload(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, err := utils.ParseTokenFromCookie(c)
	if err != nil {
		utils.RespFail(writer, "验证失败")
		return
	}
	_, isExist := models.FindUserByPhone(uc.Phone)
	if !isExist {
		utils.RespFail(writer, "用户不存在")
	}

	// 上传的文件参数
	chunkNumber, _ := strconv.Atoi(c.PostForm("chunkNumber")) // 当前分片的index
	//currentChunkSize := c.PostForm("currentChunkSize")  // 当前分片的大小，未使用
	totalSize, _ := strconv.ParseInt(c.PostForm("totalSize"), 10, 64)
	totalChunks, _ := strconv.Atoi(c.PostForm("totalChunks"))

	identifier := c.PostForm("identifier")        // 文件哈希值
	fileName := c.PostForm("filename")            // 文件名，包括拓展名
	relativePath := c.PostForm("relativePath")    // 保存的相对路径
	filePath := c.PostForm("filePath")            // 文件在用户存储区的
	isDir, _ := strconv.Atoi(c.PostForm("isDir")) // 是否是文件夹

	// filename.ext =
	split := strings.Split(fileName, ".")
	fileName = split[0]
	extendName := strings.Join(split[1:], "")

	// 保存分块文件
	uploadedFile, err := c.FormFile("file")
	err = c.SaveUploadedFile(uploadedFile, fmt.Sprintf("./repository/chunk_file/%s-%d.chunk", identifier, chunkNumber))
	if err != nil {
		log.Println(err)
		utils.RespFail(writer, "文件上传出错")
		return
	}

	if chunkNumber != totalChunks {
		utils.RespOK(writer, nil, "分块上传成功")
		return
	}

	// 走到这里意味着最后一块分块上传完成
	poolFileId := utils.GenerateUUID()
	savePath := fmt.Sprintf("../repository/upload_file/%s", poolFileId)
	// 将分块文件合并
	err = utils.MergeChunkToFile(identifier, totalChunks)
	if err != nil {
		utils.RespFail(writer, "文件上传失败")
		return
	}
	// 开始写入数据库
	ur := models.UserRepository{
		UserFileId: uc.Identity,
		UserId:     uc.Identity,
		FileId:     utils.GenerateUUID(),
		IsDir:      isDir,
		FilePath:   filePath + relativePath,
		FileName:   fileName,
		ExtendName: extendName,
		UploadTime: time.Now(),
		FileSize:   totalSize,
	}
	rp := models.RepositoryPool{
		Identity: poolFileId,
		Hash:     identifier,
		Size:     totalSize,
		Path:     savePath,
	}
	// 开启事务，插入repository_pool, user_repository
	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&ur).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		if err := tx.Create(&rp).Error; err != nil {
			// 返回任何错误都会回滚事务
			return err
		}
		// 返回 nil 提交事务
		return nil
	})
	if err != nil {
		utils.RespFail(writer, err.Error())
		return
	}
	utils.RespOK(writer, nil, "文件上传成功")
}
