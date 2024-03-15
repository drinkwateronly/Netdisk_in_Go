package models

import (
	"gorm.io/gorm"
)

type UserBasic struct {
	gorm.Model
	//Account  string `valid:"matches(^[a-zA-Z0-9]{6,}$)"` // 账号，数字或字母，6~20位
	UserId           string
	UserType         uint8
	Username         string
	Password         string
	Phone            string
	Email            string
	TotalStorageSize uint64 // 总存储量，byte为单位
	StorageSize      uint64 // 已使用存储量，byte为单位
	Salt             string
	//ClientIp      string
	//ClientPort    string
	//IsLoginOut    bool      `gorm:"column:is_login_out" json:"is_login_out"`
	//DeviceInfo    string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func FindUserByPhone(DB *gorm.DB, phone string) (*UserBasic, bool) {
	ub := UserBasic{}
	res := DB.Where("phone = ?", phone).Find(&ub)
	if res.RowsAffected == 0 || res.Error != nil { // 用户不存在
		return nil, false
	}
	return &ub, true
}

func FindUserByIdentity(db *gorm.DB, userId string) (*UserBasic, bool, error) {
	ub := UserBasic{}
	res := db.Where("user_id = ?", userId).Find(&ub)
	if res.Error != nil {
		return nil, false, res.Error
	}
	if res.RowsAffected == 0 { // 用户不存在
		return nil, false, nil
	}
	return &ub, true, nil
}
