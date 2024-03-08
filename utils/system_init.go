package utils

import (
	"gopkg.in/yaml.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"log"
	"os"
	"time"
)

var DB *gorm.DB
var MyLog *log.Logger
var Config ConfigModel

func InitLogger() {
	file, _ := os.Create("logger.txt")
	MyLog = log.New(file, "example ", log.Ldate|log.Ltime|log.Lshortfile)
}

func InitMySQL() {
	// 自定义SQL语句日志
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			Colorful:                  true,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			ParameterizedQueries:      false,
		},
	)
	var err error
	DB, err = gorm.Open(mysql.Open("root:19990414@tcp(127.0.0.1:3306)/netdisk?charset=utf8mb4&parseTime=True&loc=Local"),
		&gorm.Config{
			Logger: newLogger, // log
		})
	if err != nil {
		panic("failed to connect mysql")
	}
}

func InitConfig() error {
	f, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	var data []byte
	data, err = io.ReadAll(f)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		return err
	}
	return nil
}
