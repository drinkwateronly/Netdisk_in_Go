package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	docs "netdisk_in_go/docs"
	"netdisk_in_go/handler"
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
	r.POST("/user/register", handler.UserRegister)
	r.GET("/user/login", handler.UserLogin)
	r.GET("/user/checkuserlogininfo", handler.CheckLogin)

	// 存储
	r.GET("/filetransfer/getstorage", handler.GetUserStorage)
	r.GET("/file/getfilelist", handler.GetUserFileList)
	r.GET("/filetransfer/uploadfile", handler.FileUploadPrepare)
	r.POST("/filetransfer/uploadfile", handler.FileUpload)

	// 文件下载
	r.GET("/filetransfer/downloadfile", handler.FileDownload)
	r.GET("/filetransfer/batchDownloadFile", handler.FileDownloadInBatch)

	// 文件操作
	r.GET("/filetransfer/preview", handler.FilePreview)

	// 文件夹操作
	fileAPI := r.Group("file")
	fileAPI.Use(handler.Authentication)
	fileAPI.POST("/createFold", handler.CreateFolder)
	fileAPI.POST("/createFile", handler.CreateFile)

	fileAPI.POST("/deletefile", handler.DeleteFile)
	fileAPI.POST("/batchdeletefile", handler.DeleteFilesInBatch)
	fileAPI.POST("/renamefile", handler.RenameFile)
	fileAPI.GET("/getfiletree", handler.GetFileTree)
	fileAPI.POST("/movefile", handler.MoveFile)

	// office
	officeAPI := r.Group("office")
	officeAPI.POST("/previewofficefile", handler.PreviewOfficeFile)
	officeAPI.GET("/filedownload", handler.OfficeFileDownload)
	officeAPI.GET("/preview", handler.OfficeFilePreview)
	officeAPI.POST("/callback", handler.OfficeCallback)

	//回收站
	//recoveryAPI := r.Group("")
	r.GET("recoveryfile/list", handler.GetRecoveryFileList)
	r.POST("recoveryfile/deleterecoveryfile", handler.DelRecoveryFile)
	r.POST("recoveryfile/batchdelete", handler.DelRecoveryFilesInBatch)

	// 文件分析
	r.POST("share/sharefile", handler.ShareFiles)
	r.GET("/share/checkendtime", handler.CheckShareEndTime)
	r.GET("/share/sharetype", handler.CheckShareType)
	r.GET("/share/checkextractioncode", handler.CheckShareExtractionCode)
	r.GET("/share/sharefileList", handler.GetShareFileList)
	r.POST("share/savesharefile", handler.SaveShareFile)
	return r
}
