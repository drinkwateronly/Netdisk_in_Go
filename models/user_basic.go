package models

import (
	"gorm.io/gorm"
	"netdisk_in_go/utils"
)

type UserBasic struct {
	gorm.Model
	//Account  string `valid:"matches(^[a-zA-Z0-9]{6,}$)"` // 账号，数字或字母，6~20位
	Username         string
	Password         string
	Salt             string
	Phone            string
	Email            string
	UserIdentity     string
	UserType         uint
	TotalStorageSize int64 // 总存储量，byte为单位
	StorageSize      int64 // 已使用存储量，byte为单位
	//ClientIp      string
	//ClientPort    string
	//IsLoginOut    bool      `gorm:"column:is_login_out" json:"is_login_out"`
	//DeviceInfo    string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func CreateUser(ub *UserBasic) *gorm.DB {
	return utils.DB.Create(ub)
}

func FindUserByPhone(phone string) (*UserBasic, bool) {
	ub := &UserBasic{}
	rowsAffected := utils.DB.Where("phone = ?", phone).First(&ub).RowsAffected
	if rowsAffected == 0 {
		return nil, false
	}
	return ub, true
}
