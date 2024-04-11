package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	docs "netdisk_in_go/docs"
	"netdisk_in_go/middleware"
	"netdisk_in_go/service"
	"netdisk_in_go/service/file_service"
	"netdisk_in_go/service/office_service"
	"netdisk_in_go/service/recovery_service"
	"netdisk_in_go/service/share_service"
)

func Router() *gin.Engine {
	r := gin.Default()
	// swagger前后端分离
	// 访问：http://localhost:8080/swagger/index.html
	// 参考：https://www.jb51.net/article/259993.htm
	// 更新命令：swag config
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	docs.SwaggerInfo.BasePath = ""

	// 测试用
	r.GET("/hello", func(c *gin.Context) {
		fmt.Fprintf(c.Writer, "hello World!")
	})

	// 通用
	r.GET("/notice/list", service.NoticeList)
	r.GET("/param/grouplist", service.GetCopyright)

	// 用户
	userServiceGroup := r.Group("user")
	userServiceGroup.POST("/register", service.UserRegister)
	userServiceGroup.GET("/login", service.UserLogin)
	userServiceGroup.GET("/checkuserlogininfo", service.CheckLogin)

	// 存储
	fileTransferGroup := r.Group("filetransfer")
	fileTransferGroup.Use(middleware.Authentication)
	fileTransferGroup.GET("/getstorage", file_service.GetUserStorage)
	fileTransferGroup.GET("/uploadfile", file_service.FileUploadPrepare)
	fileTransferGroup.POST("/uploadfile", file_service.FileUpload)
	fileTransferGroup.GET("/downloadfile", file_service.FileDownload)
	fileTransferGroup.GET("/batchDownloadFile", file_service.FileDownloadInBatch)
	fileTransferGroup.GET("/preview", file_service.FilePreview)

	// 文件夹操作
	fileOpGroup := r.Group("file")
	fileOpGroup.Use(middleware.Authentication)
	fileOpGroup.GET("/getfilelist", file_service.GetUserFileList)         // 获取文件列表
	fileOpGroup.POST("/createFold", file_service.CreateFolder)            // 文件夹创建
	fileOpGroup.POST("/createFile", file_service.CreateFile)              // 文件创建
	fileOpGroup.POST("/deletefile", file_service.DeleteFile)              // 文件删除
	fileOpGroup.POST("/batchdeletefile", file_service.DeleteFilesInBatch) // 文件批量删除
	fileOpGroup.POST("/renamefile", file_service.RenameFile)              // 文件重命名
	fileOpGroup.GET("/getfiletree", file_service.GetFileTree)             // 文件移动
	fileOpGroup.POST("/movefile", func(c *gin.Context) {
		file_service.MoveFileRepost(c)
		r.HandleContext(c) // 转发请求
	})
	//fileOpGroup.POST("/movefile", file_service.MoveFile)
	fileOpGroup.POST("/batchmovefile", file_service.MoveFileInBatch) // 文件批量移动
	fileOpGroup.POST("/copyfile", file_service.CopyFile)             // 文件复制

	// office
	officeGroup := r.Group("office")
	officeGroup.POST("/previewofficefile", office_service.PreviewOfficeFile)
	officeGroup.GET("/filedownload", office_service.OfficeFileDownload)
	officeGroup.GET("/preview", office_service.OfficeFilePreview)
	officeGroup.POST("/callback", office_service.OfficeCallback)

	// 回收站
	recoveryGroup := r.Group("recoveryfile")
	recoveryGroup.Use(middleware.Authentication)
	recoveryGroup.GET("list", recovery_service.GetRecoveryFileList)            // 回收站文件列表
	recoveryGroup.POST("deleterecoveryfile", recovery_service.DelRecoveryFile) // 回收站文件删除
	recoveryGroup.POST("batchdelete", recovery_service.DelRecoveryInBatch)     // 回收站文件批量删除
	recoveryGroup.POST("restorefile", recovery_service.RestoreFile)            // 恢复回收站文件

	// 文件分享
	shareGroup := r.Group("share")
	shareGroup.Use(middleware.Authentication)
	shareGroup.GET("/sharefileList", middleware.Authentication, share_service.GetShareFileList)
	shareGroup.GET("/checkendtime", share_service.CheckShareEndTime)
	shareGroup.GET("/sharetype", share_service.CheckShareType)
	shareGroup.GET("/checkextractioncode", share_service.CheckShareExtractionCode)
	shareGroup.GET("/shareList", share_service.GetShareList)

	shareGroup.POST("/sharefile", middleware.Authentication, share_service.FilesShare)
	shareGroup.POST("/savesharefile", middleware.Authentication, share_service.SaveShareFile)

	return r
}
