package models

import (
	"archive/zip"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"netdisk_in_go/middle_models"
	"netdisk_in_go/utils"
	"os"
	"strings"
	"time"
)

// 用户存储池
type UserRepository struct {
	UserFileId    string `json:"userFileId"`
	FileId        string `json:"fileId"`
	UserId        string `json:"userId"`
	FilePath      string `json:"filePath"`
	ParentId      string `json:"parentId"`
	FileName      string `json:"fileName"`
	ExtendName    string `json:"extendName"`
	FileType      uint8  `json:"fileType"`
	IsDir         uint8  `json:"isDir"`
	FileSize      uint64 `json:"fileSize"`
	ModifyTime    string `json:"modifyTime"`
	UploadTime    string `json:"uploadTime"`
	DeleteBatchId string `json:"deleteBatchNum"`
	gorm.Model
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
func FindFilesByPathAndPage(filePath, userId string, currentPage, count uint) ([]UserRepository, int, error) {
	var files []UserRepository
	// 分页查询
	offset := count * (currentPage - 1)
	err := utils.DB.Where("user_id = ? and file_path = ?", userId, filePath).Find(&files).Error
	if err != nil {
		return nil, 0, err
	}
	// 应对最后一页时，实际记录数少于count的情况。
	if offset+count+1 > uint(len(files)) {
		return files[offset:], len(files), err
	} else {
		return files[offset : offset+count], len(files), err
	}
}

// FindFilesByTypeAndPage
// 根据当前页currentPage和每页记录count，返回分页查询的文件记录列表，并返回总记录条数（前端需要展示总的文件数量）
func FindFilesByTypeAndPage(fileType uint8, userId string, currentPage, count uint) ([]UserRepository, int, error) {
	var files []UserRepository
	// 分页查询
	offset := count * (currentPage - 1)
	// 原本使用了.Offset().Limit()，但数据库的分页查询无法获取所有记录条数
	// 获取用户对应类型的所有文件
	err := utils.DB.Where("user_id = ? and file_type = ?", userId, fileType).Find(&files).Error
	// 如果实际记录数少于count（例如最后一页）
	if offset+count >= uint(len(files)) {
		return files[offset:], len(files), err
	} else {
		return files[offset : offset+count], len(files), err
	}
}

// FindFileByNameAndPath 根据文件地址文件名，查询文件是否存在
func FindFileByNameAndPath(db *gorm.DB, userId, filePath, fileName, extendName string) (*UserRepository, bool) {
	var ur UserRepository
	rowsAffected := db.
		Where("user_id = ? and file_path = ? and file_name = ? and extend_name = ? and is_dir = 0", userId, filePath, fileName, extendName).
		Find(&ur).RowsAffected
	if rowsAffected == 0 { // 文件不存在
		return nil, false
	}
	// 文件存在或者出错
	return &ur, true

}

func FindUserFileById(userId, userFileId string) (*UserRepository, bool) {
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

func FindUserFileByIds(userId string, userFileIds []string) (*[]UserRepository, bool) {
	var file []UserRepository
	// 分页查询
	rowsAffected := utils.DB.
		Where("user_id = ? and user_file_id in ?", userId, userFileIds).
		Find(&file).RowsAffected
	if rowsAffected != int64(len(userFileIds)) { // 文件不存在
		return nil, false
	}
	// 文件存在或者出错
	return &file, true
}

func FindRepFileByUserFileId(userId, userFileId string) (*RepositoryPool, bool) {
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

func FindFilesByUserFileIds(userId string, userFileIds []string) ([]RepositoryPool, bool) {
	var rp []RepositoryPool

	rowsAffected := utils.DB.Joins("JOIN user_repository ON repository_pool.file_id = user_repository.file_id").
		Where("user_repository.user_id = ? and user_repository.user_file_id in ?", userId, userFileIds).
		Find(&rp).RowsAffected
	if rowsAffected != int64(len(userFileIds)) { // 文件不存在
		return nil, false
	}
	// 文件存在或者出错
	return rp, true
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

func GetFileTreeFromDIrV1(tx *gorm.DB, userId, userFileId, parentPath, dirName string) (*FileTreeNode, error) {
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
	err := utils.DB.Clauses(clause.Locking{Strength: "UPDATE"}).Table("(?) as user_repository", tx).
		Where("user_repository.user_id = ? AND user_repository.file_path = ? AND user_repository.is_dir = 1", userId, directoryPath).Find(&files).Error
	if err != nil {
		return nil, err
	}

	children := make([]*FileTreeNode, len(files))
	// 遍历所有子文件夹
	for i, file := range files {
		// 如果该文件是子文件夹，则进入该子文件夹删除文件
		child, err := GetFileTreeFromDIrV1(tx, userId, file.UserFileId, file.FilePath, file.FileName)
		if err != nil {
			return nil, err
		}
		children[i] = child
	}
	node.Children = children
	return &node, nil
}

// BuildFileTreeFromDIr 输入文件夹记录切片，使用层序遍历建立文件树，并返回根节点
// 此处的切片和二叉树层序遍历的切片一致 todo:有效性存疑
func BuildFileTreeFromDIr(dirs []UserRepository) *FileTreeNode {
	// 根据层序遍历构建二叉树
	// 用户一定有个根目录
	root := FileTreeNode{
		UserFileId: dirs[0].UserFileId,
		DirName:    dirs[0].FileName,
		FilePath:   dirs[0].FilePath,
		Depth:      0,
		State:      "closed",
		IsLeaf:     nil,
	}
	queue := make([]*FileTreeNode, 1)
	queue[0] = &root
	curNode := &root
	children := make([]*FileTreeNode, 0)
	dirLen := len(dirs)
	var filePath string
	// 遍历一遍所有节点
	for i := 1; i < dirLen; {
		// 当前队头作为根，找到以当前节点为父节点的节点
		if queue[0].UserFileId == dirs[i].ParentId {
			// 拼接文件路径
			if dirs[i].FilePath == "/" {
				filePath = "/" + dirs[i].FileName
			} else {
				filePath = dirs[i].FilePath + "/" + dirs[i].FileName
			}
			// 孩子节点
			child := FileTreeNode{
				UserFileId: dirs[i].UserFileId,
				DirName:    dirs[i].FileName,
				FilePath:   filePath,
				Depth:      0,
				State:      "closed",
				IsLeaf:     nil,
			}
			// 当前根节点的子树
			children = append(children, &child)
			queue = append(queue, &child)
			i++
		} else {
			// 找完了当前根的所有孩子
			// 完成链接
			curNode.Children = children
			// 当前根出队
			queue = queue[1:]
			curNode = queue[0]
			children = make([]*FileTreeNode, 0)
		}
	}
	curNode.Children = children
	return &root
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

// GetUserAllFiles 查询用户的全部文件
func GetUserAllFiles(userId string) ([]*UserRepository, error) {
	var userFiles []*UserRepository
	if err := utils.DB.Where("user_id = ?", userId).Find(&userFiles).Error; err != nil {
		return nil, err
	}
	return userFiles, nil
}

// FindParentDirFromAbsPath 从绝对路径找到父文件夹记录
// input：存放文件的文件夹的绝对路径，例如/123或者/123/456/789 或者/
// output：文件夹记录，isExist，error
func FindParentDirFromAbsPath(db *gorm.DB, userId, absPath string) (*UserRepository, bool, error) {
	var ur UserRepository
	var err error
	var res *gorm.DB
	if absPath == "/" {
		res = db.Where("user_id = ? AND file_name = '/'", userId).Find(&ur)
	} else {
		list := strings.Split(absPath[1:], "/")                   //  "/123/456/789" -> ["123","456","789"]
		folderName := list[len(list)-1]                           // ["123","456","789"] -> "456"
		folderPath := absPath[0 : len(absPath)-len(folderName)-1] // "/123/456/789"  -> "/123/456"
		if folderPath == "" {
			folderPath = "/"
		}
		res = db.Where("user_id = ? AND file_path = ? AND file_name = ? AND is_dir='1'", userId, folderPath, folderName).Find(&ur)
	}
	if res.Error != nil || res.RowsAffected == 0 {
		return nil, false, err
	}
	return &ur, true, nil
}

// FindShareFilesByPath 从分享文件路径找到文件夹记录
// input：分享文件路径filePath，分享文件shareBatchId
// output：分享文件记录切片[]ShareRepository，文件数，err
func FindShareFilesByPath(filePath, shareBatchId string) ([]ShareRepository, int, error) {
	var files []ShareRepository
	err := utils.DB.Where("share_batch_id = ? and share_file_path = ?", shareBatchId, filePath).Find(&files).Error
	if err != nil {
		return nil, 0, err
	}
	return files, len(files), err
}

// GenZipFromUserRepos 根据用户文件记录的文件拓扑生成zip压缩文件，用于文件批量/文件夹下载
// input: UserRepository切片
// output: 生成的压缩文件在服务器的存储路径，error
func GenZipFromUserRepos(userRepos ...UserRepository) (string, error) {
	// 创建一个zip压缩批量文件，使用随机名称存放
	zipUUID := utils.GenerateUUID()
	// todo:判断该随机名称文件是否存在
	zipFilePath := "./repository/zip_file/" + zipUUID + ".zip"
	zipFile, err := os.Create(zipFilePath)
	defer zipFile.Close()
	if err != nil {
		return "", err
	}

	// 创建一个zip.Writer用于写入压缩文件
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	var fileFullPath string
	// 循环所有用户文件记录
	for _, userRepo := range userRepos {
		// 找到当前文件的带路径记录 或 文件夹内所有文件的带路径记录
		var userReposWithSavePath []middle_models.UserRepoWithSavePath
		userReposWithSavePath, err = middle_models.FindUserReposWithSavePath(userRepo.UserId, userRepo.UserFileId, userRepo.IsDir)
		if err != nil {
			return "", err
		}
		// 获取当前文件（文件夹）的父级文件夹绝对路径及其长度，后续
		// 而该文件在用户存储区的路径的前半段 - 这一段长度 = 文件在zip文件里的绝对路径
		// 例如：文件路径：/123/456/789/1.txt，父文件夹：/123/456/789/，zip内文件绝对路径1.txt
		var curParentAbsPath string
		if userRepo.FilePath == "/" {
			curParentAbsPath = "/"
		} else {
			curParentAbsPath = userRepo.FilePath + "/"
		}
		rootAbsPathLen := len(curParentAbsPath)

		// 当前文件记录若是文件夹
		for _, userRepoWithPath := range userReposWithSavePath {
			// 根据下载是否是文件夹分两种情况
			// case 1：是文件夹，则在zip根据路径创建文件夹即可
			if userRepoWithPath.IsDir == 1 {
				if userRepoWithPath.FilePath == "/" {
					fileFullPath = "/" + userRepoWithPath.FileName + "/" // "/123/
				} else {
					fileFullPath = userRepoWithPath.FilePath + "/" + userRepoWithPath.FileName + "/"
				}
				folderPathInZip := fileFullPath[rootAbsPathLen:] // 去除前面的根目录长度
				// 往zip里创建文件夹
				// zipWriter.Create创建文件的规则
				// "123/" 根目录创建123文件夹
				// "123/456" 在文件夹123创建456文件
				// "123/456/" 在文件夹123创建456文件夹
				_, err := zipWriter.Create(folderPathInZip)
				if err != nil {
					return "", err
				}
				continue
			}
			// case 2：是文件，需要将文件输入到zip中
			// 获取文件信息
			_, err := os.Stat(userRepoWithPath.Path)
			// 文件不存在，返回错误信息
			if os.IsNotExist(err) {
				return "", err
			}
			// 文件存在，打开文件
			file, err := os.OpenFile(userRepoWithPath.Path, os.O_RDONLY, 0777)
			if err != nil {
				return "", err
			}

			// 先获取文件完整路径（包括文件名称）
			if userRepoWithPath.FilePath == "/" {
				fileFullPath = "/" + userRepoWithPath.FileName + "." + userRepoWithPath.ExtendName // "/123/
			} else {
				fileFullPath = userRepoWithPath.FilePath + "/" + userRepoWithPath.FileName + "." + userRepoWithPath.ExtendName
			}
			// 去掉根目录路径长度，就是存放到zip中的文件路径
			filePathInZip := fileFullPath[rootAbsPathLen:] // 去除前面的根目录长度

			// 根据该路径写入文件
			fileInZipWriter, err := zipWriter.Create(filePathInZip)
			if err != nil {
				return "", err
			}
			_, err = io.Copy(fileInZipWriter, file)
			if err != nil {
				return "", err
			}
			file.Close()
		}
	}
	return zipFilePath, nil
}
