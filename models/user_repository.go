package models

import (
	"archive/zip"
	"errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
	"io"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"os"
	"strings"
	"time"
)

// UserRepository 用户存储池
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

	// gorm.Model
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt
}

func (table UserRepository) TableName() string {
	return "user_repository"
}

// PageQueryFilesByPath
// 根据文件夹地址filePath、当前页currentPage（从0开始）、每页记录数量count、
// 返回分页查询的文件记录列表，并返回总记录条数（前端需要展示总的文件数量）
func PageQueryFilesByPath(filePath, userId string, currentPage, count uint) ([]api.UserFileListResp, int, error) {
	var files []api.UserFileListResp
	// 原本使用了.Offset().Limit()，但数据库的分页查询无法获取所有记录条数
	err := DB.Model(&UserRepository{}).Where("user_id = ? and file_path = ?", userId, filePath).Scan(&files).Error
	if err != nil {
		return nil, 0, err
	}
	// 从文件记录的offset处获取count条
	offset := count * (currentPage - 1)
	if offset+count+1 > uint(len(files)) {
		// offset处获取count条大于文件总数量（例如最后一页的记录少于count条）
		return files[offset:], len(files), err
	} else {
		return files[offset : offset+count], len(files), err
	}
}

// PageQueryFilesByType
// 根据文件夹类型fileType、当前页currentPage（从0开始）、每页记录数量count、
// 返回分页查询的文件记录列表，并返回总记录条数（前端需要展示总的文件数量）
func PageQueryFilesByType(fileType uint8, userId string, currentPage, count uint) ([]api.UserFileListResp, int, error) {
	var files []api.UserFileListResp
	// 原本使用了.Offset().Limit()，但数据库的分页查询无法获取所有记录条数
	err := DB.Model(&UserRepository{}).Where("user_id = ? and file_type = ? AND is_dir = 0", userId, fileType).Scan(&files).Error
	if err != nil {
		return nil, 0, err
	}
	// 从文件记录的offset处获取count条
	offset := count * (currentPage - 1)
	if offset+count+1 > uint(len(files)) {
		// offset处获取count条大于文件总数量（例如最后一页的记录少于count条）
		return files[offset:], len(files), err
	} else {
		return files[offset : offset+count], len(files), err
	}
}

// FindFileByNameAndPath 根据文件地址文件名，查询文件是否存在
func FindFileByNameAndPath(db *gorm.DB, userId, filePath, fileName, extendName string) (*UserRepository, bool, error) {
	var ur UserRepository
	res := db.Where("user_id = ? and file_path = ? and file_name = ? and extend_name = ? and is_dir = 0", userId, filePath, fileName, extendName).
		Find(&ur)
	if res.RowsAffected == 0 { // 文件不存在
		return nil, false, nil
	}
	// 文件存在或者出错
	return &ur, true, nil
}

func FindUserFileById(tx *gorm.DB, userId, userFileId string) (*UserRepository, error) {
	var file UserRepository
	// 分页查询
	err := tx.Where("user_id = ? and user_file_id = ?", userId, userFileId).
		First(&file).Error
	if err != nil { // 文件不存在
		return nil, err
	}
	// 文件存在或者出错
	return &file, nil
}

func FindUserFilesByIds(tx *gorm.DB, userId string, userFileIds []string) ([]*UserRepository, error) {
	var file []*UserRepository

	res := tx.Where("user_id = ? and user_file_id in ?", userId, userFileIds).
		Find(&file)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected != int64(len(userFileIds)) { // 文件不存在
		return nil, errors.New("file not exist")
	}
	// 文件存在或者出错
	return file, nil
}

// FindRepFileByUserFileId
// 通过用户文件id，联表查询其中心存储池文件记录
func FindRepFileByUserFileId(db *gorm.DB, userId, userFileId string) (*RepositoryPool, bool) {
	var rp RepositoryPool
	// 联表查询
	err := db.Joins("JOIN user_repository ON repository_pool.file_id = user_repository.file_id").
		Where("user_repository.user_id = ? and user_repository.user_file_id = ?", userId, userFileId).
		First(&rp).Error
	if err != nil { // 文件不存在
		return nil, false
	}
	return &rp, true
}

// BuildFileTree 输入用户id，根据深度优先结果建立文件树，并返回根节点
func BuildFileTree(userId string) (*api.UserFileTreeNode, error) {
	// 存放查询结果
	var dirs []UserRepository
	// 用户一定有个根目录, 从根目录递归mysql查询所有文件夹
	res := DB.Raw(`with RECURSIVE temp as
(
    SELECT * from user_repository where file_name="/" AND user_id = ? AND deleted_at = 0
    UNION ALL
    SELECT ur.* from user_repository as ur,temp t 
	where ur.parent_id=t.user_file_id and ur.is_dir = 1 AND ur.deleted_at = 0
)
select * from temp;`, userId).Find(&dirs)
	if res.Error != nil {
		return nil, res.Error
	}
	// 递归mysql查询结果中，越上层的文件记录越靠前，且同一个父文件夹下的结果都会相邻
	root := api.UserFileTreeNode{
		UserFileId: dirs[0].UserFileId,
		DirName:    dirs[0].FileName,
		FilePath:   dirs[0].FilePath,
		Depth:      0,
		State:      "closed",
		IsLeaf:     nil,
		Children:   make([]*api.UserFileTreeNode, 0),
	}
	// 建队，根节点入队
	nodeMaps := make(map[string]*api.UserFileTreeNode)
	//queue := make([]*api.UserFileTreeNode, 1)
	nodeMaps[root.UserFileId] = &root
	//children := make([]*api.UserFileTreeNode, 0)
	// 存放节点文件路径
	var filePath string
	//curParentId := root.UserFileId
	//curNode := &root
	// 遍历一遍查询结果dirs
	dirLen := len(dirs)
	for i := 1; i < dirLen; i++ {
		// 拼接文件路径
		if dirs[i].FilePath == "/" {
			filePath = "/" + dirs[i].FileName
		} else {
			filePath = dirs[i].FilePath + "/" + dirs[i].FileName
		}
		// 孩子节点
		child := api.UserFileTreeNode{
			ParentId:   dirs[i].ParentId,
			UserFileId: dirs[i].UserFileId,
			DirName:    dirs[i].FileName,
			FilePath:   filePath,
			Depth:      0,
			State:      "closed",
			IsLeaf:     nil,
			Children:   make([]*api.UserFileTreeNode, 0),
		}
		nodeMaps[dirs[i].UserFileId] = &child
		nodeMaps[child.ParentId].Children = append(nodeMaps[child.ParentId].Children, &child)
	}
	return &root, nil
}

// BuildFileTreeIn 输入用户id，根据广度优先结果建立文件树，并返回根节点，弃用
func BuildFileTreeIn(userId string) (*api.UserFileTreeNode, error) {
	// 存放查询结果
	var dirs []UserRepository
	// 用户一定有个根目录, 从根目录递归mysql查询所有文件夹
	res := DB.Raw(`with RECURSIVE temp as
(
    SELECT * from user_repository where file_name="/" AND user_id = ?
    UNION ALL
    SELECT ur.* from user_repository as ur,temp t 
	where ur.parent_id=t.user_file_id and ur.is_dir = 1 AND ur.deleted_at = 0
)
select * from temp;`, userId).Find(&dirs)
	if res.Error != nil {
		return nil, res.Error
	}
	// 递归mysql查询结果中，越上层的文件记录越靠前，且同一个父文件夹下的结果都会相邻
	root := api.UserFileTreeNode{
		UserFileId: dirs[0].UserFileId,
		DirName:    dirs[0].FileName,
		FilePath:   dirs[0].FilePath,
		Depth:      0,
		State:      "closed",
		IsLeaf:     nil,
	}
	// 建队，根节点入队
	nodeMaps := make(map[string]*api.UserFileTreeNode)
	//queue := make([]*api.UserFileTreeNode, 1)
	nodeMaps[root.UserFileId] = &root
	// 设置为当前节点，创建孩子节点空列表

	children := make([]*api.UserFileTreeNode, 0)
	// 存放节点文件路径
	var filePath string
	curParentId := root.UserFileId
	// 遍历一遍查询结果dirs
	dirLen := len(dirs)
	for i := 1; i < dirLen; {
		// 找到以当前节点为父节点的节点
		if curParentId == dirs[i].ParentId {
			// 拼接文件路径
			if dirs[i].FilePath == "/" {
				filePath = "/" + dirs[i].FileName
			} else {
				filePath = dirs[i].FilePath + "/" + dirs[i].FileName
			}
			// 孩子节点
			child := api.UserFileTreeNode{
				UserFileId: dirs[i].UserFileId,
				DirName:    dirs[i].FileName,
				FilePath:   filePath,
				Depth:      0,
				State:      "closed",
				IsLeaf:     nil,
			}
			// 当前根节点的子树
			children = append(children, &child)
			// 孩子节点入队
			nodeMaps[child.UserFileId] = &child
			// 当且仅当找到了孩子，指针才移动
			i++
		} else {
			// 找完了当前节点的所有孩子
			nodeMaps[curParentId].Children = children
			// 当前节点（队头）出队
			delete(nodeMaps, curParentId)

			curParentId = dirs[i].ParentId
			// 重置孩子节点切片
			children = make([]*api.UserFileTreeNode, 0)
		}
	}
	nodeMaps[curParentId].Children = children
	return &root, nil
}

// GetUserAllFiles 查询用户的全部文件
func GetUserAllFiles(userId string) ([]*UserRepository, error) {
	var userFiles []*UserRepository
	if err := DB.Where("user_id = ?", userId).Find(&userFiles).Error; err != nil {
		return nil, err
	}
	return userFiles, nil
}

// FindParentDirFromFilePath 从文件路径找到父文件夹记录
// input：存放文件的文件夹的文件路径，例如/123或者/123/456/789 或者/
// output：文件夹记录，isExist，error
func FindParentDirFromFilePath(db *gorm.DB, userId, filePath string) (*UserRepository, error) {
	var ur UserRepository
	var res *gorm.DB
	if filePath == "/" {
		res = db.Where("user_id = ? AND file_name = '/'", userId).First(&ur)
	} else {
		list := strings.Split(filePath[1:], "/")                    //  "/123/456/789" -> ["123","456","789"]
		folderName := list[len(list)-1]                             // ["123","456","789"] -> "456"
		folderPath := filePath[0 : len(filePath)-len(folderName)-1] // "/123/456/789"  -> "/123/456"
		if folderPath == "" {
			folderPath = "/"
		}
		res = db.Where("user_id = ? AND file_path = ? AND file_name = ? AND is_dir='1'", userId, folderPath, folderName).First(&ur)
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return &ur, nil
}

// GenZipFromUserRepos 根据用户文件记录的文件拓扑生成zip压缩文件，用于文件批量/文件夹下载
// input: UserRepository切片
// output: 生成的压缩文件在服务器的存储路径，error
func GenZipFromUserRepos(reqUserRepos ...*UserRepository) (string, error) {
	// 创建一个zip压缩批量文件，使用随机名称存放
	zipUUID := common.GenerateUUID()
	zipFilePath := "./repository/zip_file/" + zipUUID + ".zip"

	// 如果文件已存在，即UUID重复，Create会将文件清空
	zipFile, err := os.Create(zipFilePath)
	defer zipFile.Close()
	if err != nil {
		return "", err
	}

	// 创建一个zip.Writer用于写入压缩文件
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	//var fileFullPath string

	// 循环所有请求下载的用户文件记录（其中可能有文件或文件夹）
	for _, reqUserRepo := range reqUserRepos {
		var userReposWithSavePath []UserRepoWithSavePath
		// UserRepoWithSavePath即带文件保存路径的user_repository
		// 注意，如果是文件，len(userReposWithSavePath) = 1
		// 如果是文件夹，len(userReposWithSavePath) >= 1
		userReposWithSavePath, err = FindUserReposWithSavePath(reqUserRepo.UserId, reqUserRepo.UserFileId, reqUserRepo.IsDir)
		if err != nil {
			return "", err
		}

		// 随后，处理userReposWithSavePath中的所有文件路径
		// 假设当前要下载的文件夹在用户存储区的绝对路径为 "/123/456/789"
		// 那么下载的文件夹在zip文件中的绝对路径为"789"
		// 该下载文件夹的第一层级内部文件则以"789"作为父文件夹，后续层级以此类推。
		// 因此需要删除掉 【用户存储区的路径的前半段"/123/456/"】
		// 【这个前半段】就是要下载的文件夹的父文件绝对路径"/123/456" + "/"
		// 一个特殊情况是，当下载的文件在根目录下时，"/123"，只需要删除前半段的"/"即可

		var deleteLen int // 要删除的前半段路径长度
		if reqUserRepo.FilePath == "/" {
			// 特殊情况，下载文件在根目录
			deleteLen = 1
		} else {
			// 一般情况，下载的文件不在根目录，reqUserRepo.FilePath + "/"
			deleteLen = len(reqUserRepo.FilePath) + 1
		}

		for _, userRepoWithPath := range userReposWithSavePath {
			// 循环所有要下载的文件
			if userRepoWithPath.IsDir == 1 {
				// case 1：下载的文件是文件夹
				// 拼接用户文件记录中的文件路径+文件名
				fileFullPath := filehandler.ConCatFileFullPath(userRepoWithPath.FilePath, userRepoWithPath.FileName)
				// zip格式中，以"/"结尾表示文件夹
				fileFullPath += "/"
				// 得到存放到zip文件的路径
				folderPathInZip := fileFullPath[deleteLen:] // 去除前面的根目录长度
				// zipWriter.Create创建文件的规则
				// "123/" 根目录创建123文件夹
				// "123/456" 在文件夹123创建456文件
				// "123/456/" 在文件夹123创建456文件夹
				_, err := zipWriter.Create(folderPathInZip)
				// 不需要写入文件，创建文件夹即可
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
			// 拼接用户文件记录中的文件路径+文件名+拓展名
			fileFullPath := filehandler.ConCatFileFullPath(userRepoWithPath.FilePath, userRepoWithPath.FileName+"."+userRepoWithPath.ExtendName)
			// 去掉根目录路径长度，就是存放到zip中的文件路径
			filePathInZip := fileFullPath[deleteLen:]
			// 根据该路径写入zip文件
			fileInZipWriter, err := zipWriter.Create(filePathInZip)
			if err != nil {
				return "", err
			}
			_, err = io.Copy(fileInZipWriter, file)
			if err != nil {
				return "", err
			}
			file.Close() // 每次循环都要关闭文件
		}
	}
	// todo: 理论上可以直接返回
	return zipFilePath, nil
}

// FindUserReposWithSavePath 找到带文件存储地址的UserRepository
// 情况1：当前输入的用户文件id对应是文件，那么返回该文件的UserRepoWithSavePath
// 情况2：当前输入的用户文件id对应是文件夹，那么将返回该文件夹下所有文件（文件夹）的UserRepoWithSavePath切片
func FindUserReposWithSavePath(userId, userFileId string, isDir uint8) ([]UserRepoWithSavePath, error) {
	var userReposWithSavePath []UserRepoWithSavePath
	var res *gorm.DB
	if isDir == 0 { //  情况1：当前文件为非文件夹，直接联表查询该文件记录（附带其存储地址）
		res = DB.Raw(`SELECT * FROM user_repository AS ur JOIN repository_pool AS rp ON rp.file_id = ur.file_id 
WHERE ur.user_file_id= ? AND ur.user_id = ? `,
			userFileId, userId).Find(&userReposWithSavePath)
	} else { //  情况2: 当前文件为文件夹，使用递归查询
		res = DB.Raw(
			`SELECT recur.*, rp.path FROM(with RECURSIVE temp as
(
SELECT * FROM user_repository where user_file_id= ? AND user_id = ?
UNION all
SELECT ur.* FROM user_repository 
AS ur,temp t 
WHERE ur.parent_id=t.user_file_id AND ur.user_id = ? AND ur.deleted_at = 0 
)SELECT * FROM temp) AS recur LEFT JOIN repository_pool AS rp ON rp.file_id = recur.file_id`,
			userFileId, userId, userId).Find(&userReposWithSavePath)
	}
	if res.Error != nil || res.RowsAffected == 0 {
		// 出错或没有找到记录
		return nil, res.Error
	}
	return userReposWithSavePath, nil
}

// FindFolderFromAbsPath 根据文件夹绝对路径查询记录
func FindFolderFromAbsPath(tx *gorm.DB, userId, absPath string) (*UserRepository, error) {
	var file UserRepository
	if absPath == "/" {
		err := tx.Where("user_id = ? AND file_name = '/'", userId).First(&file).Error
		if err != nil {
			// err包括记录不存在
			return nil, err
		}
		return &file, nil
	}
	// 从绝对路径分离出文件夹名称及其父文件夹绝对路径
	parentPath, folderName, err := filehandler.SplitAbsPath(absPath)
	if err != nil {
		return nil, err
	}
	// 查询该文件夹
	err = tx.Where("file_path = ? AND file_name = ? AND user_id = ? AND is_dir = 1", parentPath, folderName, userId).
		First(&file).Error
	if err != nil {
		// err包括记录不存在
		return nil, err
	}
	return &file, err
}

// FindShareFilesByPathAndPage
// 根据文件夹地址filePath、当前页currentPage（从0开始）、每页记录数量count、
// 返回分页查询的文件记录列表，并返回总记录条数（前端需要展示总的文件数量）

type UserRepoWithSavePath struct {
	UserFileId string `json:"userFileId"`
	FileId     string `json:"fileId"`
	UserId     string `json:"userId"`
	FilePath   string `json:"filePath"`
	ParentId   string `json:"parentId"`
	FileName   string `json:"fileName"`
	ExtendName string `json:"extendName"`
	FileType   uint8  `json:"fileType"`
	IsDir      uint8  `json:"isDir"`
	FileSize   uint64 `json:"fileSize"`
	Path       string `json:"path"` // 文件的真实保存位置
}

// FindAllFilesFromFileId 找到某个文件夹 及 其包含的所有文件 的记录
func FindAllFilesFromFileId(tx *gorm.DB, userId, dirId string) ([]*UserRepository, error) {
	var dirs []*UserRepository
	// 只需要在递归的初始条件限定user_id即可
	err := tx.Raw(`with RECURSIVE temp as
(
    SELECT * from user_repository where user_file_id= ? AND user_id = ? AND deleted_at = 0
    UNION ALL
    SELECT ur.* from user_repository as ur,temp t 
	where ur.parent_id=t.user_file_id AND ur.deleted_at = 0
)
select * from temp;`, dirId, userId).Find(&dirs).Error
	if err != nil {
		return nil, err
	}
	if len(dirs) == 0 {
		return nil, errors.New("record not found")
	}
	return dirs, err
}

func FindRoot(tx *gorm.DB, userId string) (*UserRepository, error) {
	var root UserRepository
	err := tx.Where("user_id=? AND file_name = '/'", userId).First(&root).Error
	if err != nil {
		return nil, err
	}
	return &root, nil
}
