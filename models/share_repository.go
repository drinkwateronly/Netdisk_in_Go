package models

import (
	"errors"
	"gorm.io/gorm"
	"netdisk_in_go/common/api"
)

type ShareRepository struct {
	gorm.Model
	FileId        string `json:"fileId"`
	UserFileId    string `json:"userFileId"`
	ParentId      string `json:"parentId"`
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

// FindMyShareList 分页查询
func FindMyShareList(userId string, req api.GetShareListReq) ([]api.GetMyShareListResp, int, error) {
	filePath := req.ShareFilePath
	shareBatchId := req.ShareBatchId
	count := req.PageCount
	currentPage := req.CurrentPage

	var files []api.GetMyShareListResp
	// 原本使用了.Offset().Limit()，但数据库的分页查询无法获取所有记录条数
	var err error
	if req.ShareFilePath == "/" {
		// 不需要分享批次，展示所有分享文件
		err = DB.Table("share_basic as sb").Select("sb.*, sr.*").
			Joins("LEFT JOIN share_repository AS sr ON sb.share_batch_id = sr.share_batch_id").
			Where("sb.user_id = ? AND sr.share_file_path = ?", userId, filePath).Scan(&files).Error
	} else {
		// 需要批次，展示某批次内的文件
		err = DB.Table("share_basic as sb").Select("sb.*, sr.*").
			Joins("LEFT JOIN share_repository AS sr ON sb.share_batch_id = sr.share_batch_id").
			Where("sb.user_id = ? AND sr.share_batch_id = ? AND sr.share_file_path = ?", userId, shareBatchId, filePath).Scan(&files).Error
	}
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

func FindAllShareFilesFromFileId(tx *gorm.DB, shareBatchId, dirId string) ([]*ShareRepository, error) {
	var dirs []*ShareRepository
	// 只需要在递归的初始条件限定user_id即可
	err := tx.Raw(`with RECURSIVE temp as
(
    SELECT * from share_repository where share_batch_id = ? AND user_file_id= ? AND deleted_at IS NULL
    UNION ALL
    SELECT sr.* from share_repository as sr,temp t  
	where sr.parent_id=t.user_file_id AND sr.deleted_at IS NULL
)
select * from temp;`, shareBatchId, dirId).Find(&dirs).Error
	if err != nil {
		return nil, err
	}
	if len(dirs) == 0 {
		return nil, errors.New("record not found")
	}
	return dirs, err
}

func FindShareFilesByIds(tx *gorm.DB, userFileIds []string) ([]*ShareRepository, error) {
	var file []*ShareRepository

	res := tx.Where("user_file_id in ?", userFileIds).Find(&file)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != int64(len(userFileIds)) { // 文件不存在
		return nil, errors.New("file not exist")
	}
	// 文件存在或者出错
	return file, nil
}
