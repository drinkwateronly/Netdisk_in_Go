package models

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type UserBasic struct {
	// gorm.Model
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

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
	res := DB.Where("phone = ?", phone).First(&ub)
	if res.Error != nil { // 用户不存在
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

// UpdateUserStorageSize 更新用户可存储的剩余容量
func UpdateUserStorageSize(tx *gorm.DB, userId string, newStorageSize uint64) error {
	res := tx.Where("user_id = ?", userId).Updates(UserBasic{
		StorageSize: newStorageSize,
	})
	// 调用时应当保证用户存在，因此不处理res.RowsAffected == 0
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// GetUserBasicFromContext 从gin.Context中获取中间件存放的userBasic字段，如果没有，则是未登录状态
func GetUserBasicFromContext(c *gin.Context) (*UserBasic, bool) {
	ub, boo := c.Get("userBasic")
	if !boo {
		return nil, false
	}
	return ub.(*UserBasic), true
}
