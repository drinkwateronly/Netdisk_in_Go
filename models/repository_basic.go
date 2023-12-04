package models

import (
	"netdisk_in_go/utils"
	"time"
)

// 中心存储池
type RepositoryPool struct {
	Id        int64
	Identity  string
	Hash      string
	Size      int64
	Path      string
	CreatedAt time.Time `gorm:"created"`
	UpdatedAt time.Time `gorm:"updated"`
	DeletedAt time.Time `gorm:"deleted"`
}

func (RepositoryPool) TableName() string {
	return "repository_pool"
}

// 用户存储池
type UserRepository struct {
	Id         int64     `json:"id"`
	UserFileId string    `json:"userFileId"`
	UserId     string    `json:"userId"`
	FileId     string    `json:"fileId"`
	IsDir      int       `json:"isDir"`
	FilePath   string    `json:"filePath"`
	FileName   string    `json:"fileName"`
	ExtendName string    `json:"extendName"`
	UploadTime time.Time `json:"uploadTime"`
	FileSize   int64     `json:"fileSize"`
	//UpdatedAt  time.Time `gorm:"updated"`
	//DeletedAt  time.Time `gorm:"deleted"`
}

func (table UserRepository) TableName() string {
	return "user_repository"
}

func FindFilesByPath(path, userIdentity string, fileType, currentPage, pageCount int) ([]UserRepository, error) {
	var files []UserRepository
	// 分页查询
	offset := pageCount * (currentPage - 1)
	err := utils.DB.Where("file_path = ?", path).Find(&files).
		Offset(offset).Limit(pageCount).Error
	return files, err
}
