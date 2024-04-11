package models

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"time"
)

// 回收站文件
type RecoveryBatch struct {
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
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (RecoveryBatch) TableName() string {
	return "recovery_batch"
}

// SoftDelUserFiles 根据userId, userFileId，将单个/多个用户文件记录软删除，并为记录设置delBatchId
func SoftDelUserFiles(tx *gorm.DB, delTime time.Time, delBatchId, userId string, userFiles ...*UserRepository) (uint64, error) {
	userFileIds := make([]string, len(userFiles))
	var delStorage uint64
	for i, userFile := range userFiles {
		userFileIds[i] = userFile.UserFileId // 删除的文件id
		delStorage += userFile.FileSize      // 删除的文件总大小
	}
	err := tx.Where("user_id = ? and user_file_id in ?", userId, userFileIds).
		Updates(&UserRepository{
			DeletedAt:     soft_delete.DeletedAt(delTime.Unix()),
			DeleteBatchId: delBatchId, // 设置delBatchId
			//Model: gorm.Model{ // 软删除旧版本，deleted_at为时间而非时间戳
			//	DeletedAt: gorm.DeletedAt{
			//		Time:  time.Now(),
			//		Valid: true,
			//	},
			//},
		}).Error
	return delStorage, err
}

// InsertToRecoveryBatch 插入recovery_batch表
func InsertToRecoveryBatch(tx *gorm.DB, delTime time.Time, delBatchId string, urs ...*UserRepository) error {
	rbs := make([]RecoveryBatch, len(urs))
	for i := range urs {
		rbs[i] = RecoveryBatch{
			UserFileId:    urs[i].UserFileId,
			UserId:        urs[i].UserId,
			DeleteBatchId: delBatchId,
			FilePath:      urs[i].FilePath,
			FileName:      urs[i].FileName,
			FileType:      urs[i].FileType,
			ExtendName:    urs[i].ExtendName,
			IsDir:         urs[i].IsDir,
			FileSize:      urs[i].FileSize,
			DeleteTime:    delTime.Format("2006-01-02 15:04:05"),
			UploadTime:    urs[i].UploadTime,
		}
	}
	return tx.Create(&rbs).Error
}
