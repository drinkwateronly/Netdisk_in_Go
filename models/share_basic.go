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

func FindShareFilesByPathAndPage(userId string, req api.GetShareListReq) ([]api.GetShareListResp, int, error) {
	// filePath := req.ShareFilePath // 忽略filePath
	count := req.PageCount
	currentPage := req.CurrentPage

	var files []api.GetShareListResp
	// 原本使用了.Offset().Limit()，但数据库的分页查询无法获取所有记录条数
	err := DB.Table("share_basic as sb").Select("ur.*, sb.*").
		Joins("LEFT JOIN share_repository AS sr ON sb.share_batch_id = sr.share_batch_id").
		Joins("LEFT JOIN user_repository AS ur ON sr.user_file_id = ur.user_file_id").
		Where("ur.user_id = ? AND sr.share_file_path = '/'", userId).Scan(&files).Error
	if err != nil {
		return nil, 0, err
	}
	// 从所有符合条件的文件记录的offset处获取count条
	offset := count * (currentPage - 1)
	if offset+count+1 > uint(len(files)) {
		// offset处获取count条大于文件总数量（例如最后一页的记录少于count条）
		return files[offset:], len(files), err
	} else {
		return files[offset : offset+count], len(files), err
	}
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
