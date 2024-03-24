package models

import (
	"gorm.io/gorm"
	"time"
)

// 中心存储池
type RepositoryPool struct {
	//gorm.Model
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	FileId    string
	Hash      string
	Size      uint64
	Path      string
}

func (RepositoryPool) TableName() string {
	return "repository_pool"
}

func FindFileByMD5(hash string) (*RepositoryPool, bool) {
	rp := RepositoryPool{}
	rowsAffected := DB.Where("hash = ?", hash).Find(&rp).RowsAffected
	if rowsAffected == 0 { // 文件不存在
		return nil, false
	}
	return &rp, true
}
