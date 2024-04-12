package models

import (
	"gorm.io/gorm"
	"netdisk_in_go/common/api"
	"time"
)

type ShareBasic struct {
	// gorm.Model
	// gorm.Model
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserId         string    `json:"userId"`
	Salt           string    `json:"salt"`
	ShareBatchId   string    `json:"shareBatchNum"`
	ShareType      uint8     `json:"shareType"`
	ExtractionCode string    `json:"extractionCode"`
	ExpireTime     time.Time `json:"expireTime"`
}

func (ShareBasic) TableName() string {
	return "share_basic"
}

// FindShareFilesByPath 从分享文件路径找到文件夹记录
// input：分享文件路径filePath，分享文件shareBatchId
// output：分享文件记录切片[]ShareRepository，文件数，err
func FindShareFilesByPath(filePath, shareBatchId string) ([]api.GetShareFileListResp, error) {
	var files []api.GetShareFileListResp
	err := DB.Where("share_batch_id = ? and share_file_path = ?", shareBatchId, filePath).Scan(&files).Error
	if err != nil {
		return nil, err
	}
	return files, err
}
