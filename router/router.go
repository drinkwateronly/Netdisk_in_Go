package router

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/service"
)

func Router() *gin.Engine {
	r := gin.Default()
	// 用户
	r.GET("/user/login", service.UserLogin)
	r.POST("/user/register", service.UserRegister)
	r.GET("/user/checkuserlogininfo", service.CheckLogin)

	// 存储
	r.GET("/filetransfer/getstorage", service.GetUserStorage)
	r.GET("/file/getfilelist", service.GetUserFileList)
	r.GET("/filetransfer/uploadfile", service.PrepareFileUpload)
	r.POST("/filetransfer/uploadfile", service.FileUpload)
	return r
}
