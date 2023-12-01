package router

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/service"
)

func Router() *gin.Engine {
	r := gin.Default()
	r.GET("/user/login", service.UserLogin)
	r.POST("/user/register", service.UserRegister)
	return r
}
