package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"netdisk_in_go/models"
	"netdisk_in_go/router"
	"netdisk_in_go/sysconfig"
	"os"
)

var MyLog *log.Logger

func InitLogger() {
	file, _ := os.Create("logger.txt")
	MyLog = log.New(file, "example ", log.Ldate|log.Ltime|log.Lshortfile)
}

func SystemInit() error {
	sysconfig.InitConfig()
	models.InitMysql()
	InitLogger()
	return nil
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	err := SystemInit()
	if err != nil {
		panic(err)
	}
	r := router.Router()
	f, _ := os.Create("logger.txt")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	err = r.Run(":8080")
	if err != nil {
		panic("system run on 8080 failed")
	}
}
