package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"netdisk_in_go/utils"
	"time"
)

// 用户存储池
type UserRepository struct {
	gorm.Model
	UserFileId    string `json:"userFileId"`
	UserId        string `json:"userId"`
	FileId        string `json:"fileId"`
	FilePath      string `json:"filePath"`
	FileName      string `json:"fileName"`
	FileType      int    `json:"fileType"`
	DeleteBatchId string `json:"deleteBatchNum"`
	ExtendName    string `json:"extendName"`
	IsDir         int    `json:"isDir"`
	FileSize      int64  `json:"fileSize"`
	ModifyTime    string `json:"modifyTime"`
	UploadTime    string `json:"uploadTime"`
}

type FileTreeNode struct {
	UserFileId string          `json:"id"`
	DirName    string          `json:"label"`
	FilePath   string          `json:"filePath"`
	Depth      int             `json:"depth"`
	State      string          `json:"state"`
	IsLeaf     interface{}     `json:"isLeaf"`
	IconClass  string          `json:"iconClass"`
	Children   []*FileTreeNode `json:"children"`
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

// FindFileByNameAndPath 根据文件地址文件名，查询文件是否存在
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

// FindFilesByTypeAndPage
// 根据当前页currentPage和每页记录count，返回分页查询的文件记录列表，并返回总记录条数（前端需要展示总的文件数量）
func FindFilesByTypeAndPage(fileType int, userId string, currentPage, count int) ([]UserRepository, int, error) {
	var files []UserRepository
	// 分页查询
	offset := count * (currentPage - 1)
	// 原本使用了.Offset().Limit()，进行数据库的分页查询，但无法获取所有记录条数
	// 获取用户对应类型的所有文件
	err := utils.DB.Where("user_id = ? and file_type = ?", userId, fileType).Find(&files).Error
	// 应对最后一页时，实际记录数少于count的情况。
	if offset+count >= len(files)-1 {
		return files[offset:], len(files), err
	} else {
		return files[offset : offset+count], len(files), err
	}
}

// DelAllFilesFromDir 根据用户的id，文件夹所在的文件夹路径，文件夹名称，递归删除文件夹内的所有文件
func DelAllFilesFromDir(delBatchId, userId, parentPath, dirName string) error {
	var directoryPath string // 文件夹路径
	// todo: 递归sql
	// 拼接出这个文件夹的路径
	if parentPath == "/" {
		directoryPath = parentPath + dirName
	} else {
		directoryPath = parentPath + "/" + dirName
	}
	// 找到这个文件夹下的所有子文件，加排他锁
	var files []UserRepository
	err := utils.DB.Clauses(
		clause.Locking{
			Strength: "UPDATE",
		},
	).
		Where("user_id = ? AND file_path = ?", userId, directoryPath).Find(&files).Error
	if err != nil {
		return err
	}
	//
	userFileIds := make([]string, len(files))
	// 遍历所有子文件夹
	for i, file := range files {
		// 如果该文件是子文件夹，则进入该子文件夹删除文件
		if file.FileType == utils.DIRECTORY {
			err = DelAllFilesFromDir(delBatchId, userId, file.FilePath, file.FileName)
			if err != nil {
				return err
			}
		}
		// 记录该文件夹下的文件id
		userFileIds[i] = file.UserFileId
	}
	// 开始删除该文件夹下的所有文件
	err = SoftDelUserFiles(delBatchId, userId, userFileIds...)
	if err != nil {
		return err
	}
	return nil
}

func GetFileTreeFromDIr(userId, userFileId, parentPath, dirName string) (*FileTreeNode, error) {
	var directoryPath string // 文件夹路径
	// todo: 递归sql
	// 拼接出这个文件夹的路径
	if parentPath == "/" {
		directoryPath = parentPath + dirName
	} else if parentPath == "" {
		directoryPath = "/"
	} else {
		directoryPath = parentPath + "/" + dirName
	}
	// 找到这个文件夹下的所有子文件夹，加排他锁
	node := FileTreeNode{
		UserFileId: userFileId,
		DirName:    dirName,
		FilePath:   directoryPath,
		Depth:      0,
		State:      "closed",
		IsLeaf:     nil,
	}
	var files []UserRepository
	err := utils.DB.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND file_path = ? AND is_dir = 1", userId, directoryPath).Find(&files).Error
	if err != nil {
		return nil, err
	}

	children := make([]*FileTreeNode, len(files))
	// 遍历所有子文件夹
	for i, file := range files {
		// 如果该文件是子文件夹，则进入该子文件夹删除文件
		child, err := GetFileTreeFromDIr(userId, file.UserFileId, file.FilePath, file.FileName)
		if err != nil {
			return nil, err
		}
		children[i] = child
	}
	node.Children = children
	return &node, nil
}

// SoftDelUserFiles 根据userId, userFileId，将单个/多个用户文件记录软删除，并为记录设置delBatchId
func SoftDelUserFiles(delBatchId, userId string, userFileIds ...string) error {
	err := utils.DB.Where("user_id = ? and user_file_id in ?", userId, userFileIds).
		Updates(&UserRepository{
			Model: gorm.Model{ // 软删除
				DeletedAt: gorm.DeletedAt{
					Time:  time.Now(),
					Valid: true,
				},
			},
			DeleteBatchId: delBatchId, // 设置delBatchId
		}).Error
	return err
}
