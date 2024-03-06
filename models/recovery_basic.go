package models

import (
	"gorm.io/gorm"
	"netdisk_in_go/utils"
	"time"
)

// 回收站文件
type RecoveryBasic struct {
	gorm.Model
	UserFileId    string `json:"userFileId"`
	UserId        string `json:"userId"`
	DeleteBatchId string `json:"deleteBatchNum"`
	FilePath      string `json:"filePath"`
	FileName      string `json:"fileName"`
	FileType      uint8  `json:"fileType"`
	ExtendName    string `json:"extendName"`
	IsDir         uint8  `json:"isDir"`
	FileSize      uint64 `json:"fileSize"`
	DeleteTime    string `json:"deleteTime"`
	UploadTime    string `json:"uploadTime"`
}

func (RecoveryBasic) TableName() string {
	return "recovery_batch"
}

func AddFileToRecoveryBatch(ur *UserRepository, delBatchId string) error {
	rb := RecoveryBasic{
		UserFileId:    ur.UserFileId,
		UserId:        ur.UserId,
		DeleteBatchId: delBatchId,
		FilePath:      ur.FilePath,
		FileName:      ur.FileName,
		FileType:      ur.FileType,
		ExtendName:    ur.ExtendName,
		IsDir:         ur.IsDir,
		FileSize:      ur.FileSize,
		DeleteTime:    time.Now().Format("2006-01-02 15:04:05"),
		UploadTime:    ur.UploadTime,
	}
	return utils.DB.Create(&rb).Error
}
