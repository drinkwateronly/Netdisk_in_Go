package models

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"netdisk_in_go/sysconfig"
	"os"
	"time"
)

var DB *gorm.DB

// InitMysql 连接mysql并初始化配置
func InitMysql() {
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
	// 获取mysql的配置项
	conf := sysconfig.Config.MysqlConfig
	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User, conf.Pwd, conf.Host, conf.Port, conf.DBName)

	DB, err = gorm.Open(mysql.Open(dsn),
		&gorm.Config{
			Logger: newLogger, // log
		})
	if err != nil {
		panic(fmt.Sprintf("failed to connect mysql: %v", dsn))
	}
}
