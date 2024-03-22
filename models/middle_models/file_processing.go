package middle_models

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
