package api_models

type UserFileListReqAPI struct {
	FilePath    string `form:"filePath"`
	FileType    uint8  `form:"fileType"`
	CurrentPage uint   `form:"currentPage"`
	PageCount   uint   `form:"pageCount"`
}

type UserFileListRespAPI struct {
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
}

type UserFileTreeNode struct {
	ParentId   string              `json:"parentId"`
	UserFileId string              `json:"id"`
	DirName    string              `json:"label"`
	FilePath   string              `json:"filePath"`
	Depth      int                 `json:"depth"`
	State      string              `json:"state"`
	IsLeaf     interface{}         `json:"isLeaf"`
	IconClass  string              `json:"iconClass"`
	Children   []*UserFileTreeNode `json:"children"`
}

// CreateFileReqAPI 创建文件请求API
type CreateFileReqAPI struct {
	FileName   string `json:"fileName"`
	FilePath   string `json:"filePath"`
	ExtendName string `json:"extendName"`
}

type CreateFolderRequest struct {
	FolderName string `json:"fileName"`
	FolderPath string `json:"filePath"`
}

// FileUploadReqAPI
// Create or update API information request | 创建或更新API信息
// swagger:model FileUploadReqAPI
type FileUploadReqAPI struct {
	// 分片号
	ChunkNumber uint `form:"chunkNumber"  binding:"gte=0"`
	// 分片尺寸
	CurrentChunkSize uint `form:"currentChunkSize" binding:"gte=0"`
	// 分片数量
	TotalChunks uint `form:"totalChunks"  binding:"gte=0"`
	// 文件总大小
	TotalSize uint64 `form:"totalSize" binding:"gte=0"`
	// 文件哈希
	FileMD5 string `form:"identifier"`
	// 文件全名（文件名+拓展名）
	FileFullName string `form:"filename"`
	// 文件存储路径
	FilePath string `form:"filePath"`
	// 文件存储的相对路径
	RelativePath string `form:"relativePath"`
	// 文件夹，0则不是文件夹，1是文件夹。
	IsDir uint8 `form:"isDir"`
}

// FileShareReq
// 文件分享请求API
type FileShareReq struct {
	EndTime     string `form:"endTime"`     // 分享结束时长
	Remark      string `form:"remark"`      // 未使用
	ShareType   uint8  `form:"shareType"`   // 分享类型，有验证码时为1
	UserFileIds string `form:"userFileIds"` // 分享的用户文件ids
}

type SaveShareReq struct {
	FilePath      string `form:"filePath"`
	UserFileIds   string `form:"userFileIds"`
	ShareBatchNum string `form:"shareBatchNum"`
}

// MoveFileReqAPI 文件移动请求API
type MoveFileReqAPI struct {
	FilePath   string `json:"filePath"`
	UserFileId string `json:"userFileId"`
}

// RenameFileRequest 文件重命名请求API
type RenameFileRequest struct {
	FileName   string `json:"fileName"`
	UserFileId string `json:"userFileId"`
}
