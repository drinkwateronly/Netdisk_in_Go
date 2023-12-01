package test

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"netdisk_in_go/models"
)

func main() {
	db, err := gorm.Open(mysql.Open("root:19990414@tcp(127.0.0.1:3306)/netdisk?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// 迁移 schema
	db.AutoMigrate(&models.UserBasic{})

	//user := &models.UserBasic{
	//	Name: "chenjie",
	//}
	//// Create
	//db.Create(user)
	//
	//fmt.Println(db.First(&user, 1))
	//
	//// Update - 将 product 的 price 更新为 200
	//db.Model(&user).Update("Password", "990414")
	//// Update - 更新多个字段
}
