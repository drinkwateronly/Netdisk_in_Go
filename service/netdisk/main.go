package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	common "netdisk_in_go/common/logger"
	"netdisk_in_go/models"
	"netdisk_in_go/router"
	"netdisk_in_go/sysconfig"
	"os"
)

func SystemInit() {
	sysconfig.InitConfig()
	models.InitMysql()
	common.InitLogger()
}

func main() {
	go func() {
		// for pprof
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	SystemInit()

	r := router.Router()
	f, err := os.Create("logger.txt")
	if err != nil {
		panic("logger create error")
	}
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	port := sysconfig.Config.NetDiskConfig.Port
	err = r.Run(":" + port)
	if err != nil {
		panic(fmt.Sprintf("system run on %s failed", port))
	}
}
