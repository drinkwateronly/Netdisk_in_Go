package models

import (
	"gorm.io/gorm"
)

type ShareRepository struct {
	gorm.Model
	UserFileId    string `json:"userFileId"`
	ShareBatchId  string `json:"shareBatchNum"`
	ShareFilePath string `json:"shareFilePath"`
	FileName      string `json:"fileName"`
	ExtendName    string `json:"extendName"`
	FileSize      uint64 `json:"fileSize"`
	FileType      uint8  `json:"fileType"`
	IsDir         uint8  `json:"isDir"`
}

func (ShareRepository) TableName() string {
	return "share_repository"
}
