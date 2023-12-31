package router

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/service"
)

func Router() *gin.Engine {
	r := gin.Default()
	// 通用
	r.GET("/helloworld", func(context *gin.Context) {
		context.Writer.Write([]byte("/helloworld"))
	})

	r.GET("/notice/list", service.NoticeList)
	r.GET("/param/grouplist", service.GetCopyright)

	// 用户
	r.GET("/user/login", service.UserLogin)
	r.POST("/user/register", service.UserRegister)
	r.GET("/user/checkuserlogininfo", service.CheckLogin)

	// 存储
	r.GET("/filetransfer/getstorage", service.GetUserStorage)
	r.GET("/file/getfilelist", service.GetUserFileList)
	r.GET("/filetransfer/uploadfile", service.FileUploadPrepare)
	r.POST("/filetransfer/uploadfile", service.FileUpload)

	// 下载
	r.GET("/filetransfer/downloadfile", service.FileDownload)

	// 文件操作
	r.GET("/filetransfer/preview", service.FilePreview)

	// 文件夹操作
	fileAPI := r.Group("file")
	fileAPI.Use(service.Authentication)
	fileAPI.POST("/createFold", service.CreateFolder)
	fileAPI.POST("/createFile", service.CreateFile)
	fileAPI.POST("/deletefile", service.DeleteFile)
	fileAPI.POST("/batchdeletefile", service.DeleteFilesInBatch)

	// office
	officeAPI := r.Group("office")
	officeAPI.POST("/previewofficefile", service.PreviewOfficeFile)
	officeAPI.GET("/filedownload", service.OfficeFileDownload)
	officeAPI.GET("/preview", service.OfficeFilePreview)
	officeAPI.POST("/callback", service.OfficeCallback)
	return r

}
