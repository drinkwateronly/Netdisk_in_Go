package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"netdisk_in_go/router"
	"netdisk_in_go/utils"
	"os"
)

func main() {
	utils.InitMySQL()
	utils.InitLogger()
	err := utils.InitConfig()
	if err != nil {
		panic(err)
		return
	}
	r := router.Router()
	f, _ := os.Create("logger.txt")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	err = r.Run(":8080")
	if err != nil {
		panic("system run on 8080 failed")
		return
	}
}
