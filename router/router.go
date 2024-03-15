package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	docs "netdisk_in_go/docs"
	"netdisk_in_go/handler"
	"netdisk_in_go/middleware"
)

func Router() *gin.Engine {
	r := gin.Default()
	// swagger前后端分离
	// 访问：http://localhost:8080/swagger/index.html
	// 更新命令：swag init
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	docs.SwaggerInfo.BasePath = ""

	// 通用
	r.GET("/notice/list", handler.NoticeList)
	r.GET("/param/grouplist", handler.GetCopyright)

	// 用户
	userService := r.Group("user")
	userService.POST("/register", handler.UserRegister)
	userService.GET("/login", handler.UserLogin)
	userService.GET("/checkuserlogininfo", handler.CheckLogin)

	// 存储
	fileTransfer := r.Group("filetransfer")
	fileTransfer.Use(middleware.Authentication)
	fileTransfer.GET("/getstorage", handler.GetUserStorage)
	fileTransfer.GET("/uploadfile", handler.FileUploadPrepare)
	fileTransfer.POST("/uploadfile", handler.FileUpload)
	fileTransfer.GET("/downloadfile", handler.FileDownload)
	fileTransfer.GET("/batchDownloadFile", handler.FileDownloadInBatch)
	fileTransfer.GET("/preview", handler.FilePreview)

	// 文件夹操作
	fileOperation := r.Group("file")
	fileOperation.Use(middleware.Authentication)
	fileOperation.GET("/getfilelist", handler.GetUserFileList)
	fileOperation.POST("/createFold", handler.CreateFolder)
	fileOperation.POST("/createFile", handler.CreateFile)
	fileOperation.POST("/deletefile", handler.DeleteFile)
	fileOperation.POST("/batchdeletefile", handler.DeleteFilesInBatch)
	fileOperation.POST("/renamefile", handler.RenameFile)
	fileOperation.GET("/getfiletree", handler.GetFileTree)
	fileOperation.POST("/movefile", handler.MoveFile)

	// office
	officeService := r.Group("office")
	officeService.Use(middleware.Authentication)
	officeService.POST("/previewofficefile", handler.PreviewOfficeFile)
	officeService.GET("/filedownload", handler.OfficeFileDownload)
	officeService.GET("/preview", handler.OfficeFilePreview)
	officeService.POST("/callback", handler.OfficeCallback)

	// 回收站
	recoveryService := r.Group("recovery")
	recoveryService.Use(middleware.Authentication)
	recoveryService.GET("list", handler.GetRecoveryFileList)
	recoveryService.POST("deleterecoveryfile", handler.DelRecoveryFile)
	recoveryService.POST("batchdelete", handler.DelRecoveryFilesInBatch)

	// 文件分享
	shareService := r.Group("share")
	shareService.POST("/sharefile", handler.ShareFiles)
	shareService.GET("/checkendtime", handler.CheckShareEndTime)
	shareService.GET("/sharetype", handler.CheckShareType)
	shareService.GET("/checkextractioncode", handler.CheckShareExtractionCode)
	shareService.GET("/sharefileList", handler.GetShareFileList)
	shareService.POST("/savesharefile", handler.SaveShareFile)
	return r
}
