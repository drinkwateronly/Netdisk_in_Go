package test

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
	"testing"
)

func TestGorm(t *testing.T) {
	db, err := gorm.Open(mysql.Open("root:19990414@tcp(127.0.0.1:3306)/netdisk?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	// 迁移 schema
	db.AutoMigrate(&models.UserBasic{})
	db.AutoMigrate(&models.UserRepository{})
	db.AutoMigrate(&models.RepositoryPool{})
	db.AutoMigrate(&models.RecoveryBasic{})
	db.AutoMigrate(&models.ShareRepository{})
	db.AutoMigrate(&models.ShareBasic{})

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

func TestFind(t *testing.T) {
	utils.InitMySQL()
	ub, isExist := models.FindUserByPhone(utils.DB, "18927841103")
	if isExist {
		t.Fatal("?")
	}
	fmt.Println(ub)
}

func TestDigui(t *testing.T) {
	utils.InitMySQL()
	var ur []models.UserRepository
	res := utils.DB.Raw(`with RECURSIVE temp as
(
    select * from user_repository where file_name="/"
    union all
    select ur.* from user_repository as ur,temp t where ur.parent_id=t.user_file_id and ur.is_dir = 1 AND ur.deleted_at is NULL
)
select temp.parent_id, temp.user_file_id, temp.is_dir, temp.file_name from temp;`).Find(&ur)
	if res.Error != nil {
		t.Fatal(res.Error)
		return
	}
	fmt.Println(res.RowsAffected, len(ur))

}
