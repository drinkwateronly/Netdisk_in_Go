package models

import (
	"gorm.io/gorm"
	"netdisk_in_go/utils"
	"time"
)

// 用户存储池
type UserRepository struct {
	gorm.Model
	Id         int64     `json:"id"`
	UserFileId string    `json:"userFileId"`
	UserId     string    `json:"userId"`
	FileId     string    `json:"fileId"`
	IsDir      int       `json:"isDir"`
	FilePath   string    `json:"filePath"`
	FileName   string    `json:"fileName"`
	FileType   int       `json:"fileType"`
	ExtendName string    `json:"extendName"`
	UploadTime time.Time `json:"uploadTime"`
	FileSize   int64     `json:"fileSize"`
}

func (table UserRepository) TableName() string {
	return "user_repository"
}

// FindFilesByPathAndPage 根据文件地址，分页查询多个文件
func FindFilesByPathAndPage(filePath, userId string, currentPage, pageCount int) ([]UserRepository, error) {
	var files []UserRepository
	// 分页查询
	offset := pageCount * (currentPage - 1)
	err := utils.DB.Where("user_id = ? and file_path = ?", userId, filePath).Find(&files).
		Offset(offset).Limit(pageCount).Error
	return files, err
}

// FindFileByPathAndName 根据文件地址文件名，查询文件是否存在
func FindFileByNameAndPath(userId, filePath, fileName, extendName string) (*UserRepository, bool) {
	var ur UserRepository
	rowsAffected := utils.DB.
		Where("user_id = ? and file_path = ? and file_name = ? and extend_name = ?", userId, filePath, fileName, extendName).
		Find(&ur).RowsAffected
	if rowsAffected == 0 { // 文件不存在
		return nil, false
	}
	// 文件存在或者出错
	return &ur, true
}

func FindFileById(userId, userFileId string) (*UserRepository, bool) {
	var file UserRepository
	// 分页查询
	rowsAffected := utils.DB.
		Where("user_id = ? and user_file_id = ?", userId, userFileId).
		Find(&file).RowsAffected
	if rowsAffected == 0 { // 文件不存在
		return nil, false
	}
	// 文件存在或者出错
	return &file, true
}

func FindFileSavePathById(userId, userFileId string) (*RepositoryPool, bool) {
	var rp RepositoryPool
	// 分页查询
	rowsAffected := utils.DB.Joins("JOIN user_repository ON repository_pool.file_id = user_repository.file_id").
		Where("user_repository.user_id = ? and user_repository.user_file_id = ?", userId, userFileId).
		Find(&rp).RowsAffected
	if rowsAffected == 0 { // 文件不存在
		return nil, false
	}
	// 文件存在或者出错
	return &rp, true
}

// FindFilesByTypeAndPage 根据文件类型，查询所有文件
func FindFilesByTypeAndPage(fileType int, userId string, currentPage, pageCount int) ([]UserRepository, error) {
	var files []UserRepository
	// 分页查询
	offset := pageCount * (currentPage - 1)
	err := utils.DB.Where("user_id = ? and file_type = ?", userId, fileType).Find(&files).
		Offset(offset).Limit(pageCount).Error
	return files, err
}
