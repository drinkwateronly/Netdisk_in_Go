package middle_models

import (
	"gorm.io/gorm"
	"netdisk_in_go/utils"
)

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

// FindUserReposWithSavePath 找到带文件存储地址的UserRepository
// 情况1：当前输入的用户文件id对应是文件，那么返回该文件的UserRepoWithSavePath
// 情况2：当前输入的用户文件id对应是文件夹，那么将返回该文件夹下所有文件（文件夹）的UserRepoWithSavePath切片
func FindUserReposWithSavePath(userId, userFileId string, isDir uint8) ([]UserRepoWithSavePath, error) {
	var filesWithSavePath []UserRepoWithSavePath
	var res *gorm.DB
	if isDir == 0 {
		//  情况1：
		res = utils.DB.Raw(`SELECT * FROM user_repository AS ur JOIN repository_pool AS rp ON rp.file_id = ur.file_id 
WHERE ur.user_file_id= ? AND ur.user_id = ? `,
			userFileId, userId).Find(&filesWithSavePath)
	} else {
		//  情况2:
		res = utils.DB.Raw(
			`SELECT recur.*, rp.path FROM(with RECURSIVE temp as
(
SELECT * FROM user_repository where user_file_id= ? AND user_id = ?
UNION all
SELECT ur.* FROM user_repository 
AS ur,temp t 
WHERE ur.parent_id=t.user_file_id AND ur.user_id = ? AND ur.deleted_at is NULL 
)SELECT * FROM temp) AS recur LEFT JOIN repository_pool AS rp ON rp.file_id = recur.file_id`,
			userFileId, userId, userId).Find(&filesWithSavePath)
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return filesWithSavePath, nil
}
