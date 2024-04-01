package api

// ----------------------------------------------------------------

// UserStorageResp 用户存储空间响应
type UserStorageResp struct {
	StorageSize      uint64 `json:"storageSize"`
	TotalStorageSize uint64 `json:"totalStorageSize"`
}

// ----------------------------------------------------------------

// UserFileListReq 用户分页与分类查询文件列表请求参数
type UserFileListReq struct {
	FilePath    string `form:"filePath"`    // 文件夹路径
	FileType    uint8  `form:"fileType"`    // 文件类型
	CurrentPage uint   `form:"currentPage"` // 第页号
	PageCount   uint   `form:"pageCount"`   // 每页数量
}

// UserFileListResp 用户文件列表响应，用于前端展示文件信息
type UserFileListResp struct {
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

// ----------------------------------------------------------------

// RenameFileReq 文件重命名请求API
type RenameFileReq struct {
	FileName   string `json:"fileName"`   // 新文件名
	UserFileId string `json:"userFileId"` // 要修改的文件id
}

// MoveFileReq 文件移动请求API
type MoveFileReq struct {
	FilePath   string `json:"filePath"`   // 目标文件夹绝对路径
	UserFileId string `json:"userFileId"` // 源文件的用户文件标识符
}

// MoveFileInBatchReq 文件批量移动请求API
type MoveFileInBatchReq struct {
	FilePath    string `json:"filePath"`    // 目标文件夹绝对路径
	UserFileIds string `json:"userFileIds"` // 源文件的用户文件标识符
}

// ----------------------------------------------------------------

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

// ----------------------------------------------------------------

// FileDownloadReq 单个文件下载请求
type FileDownloadReq struct {
	UserFileId string `form:"userFileId"` // 单个下载文件的用户文件标识符
}

// FileDownloadInBatchReq 文件批量下载请求
type FileDownloadInBatchReq struct {
	UserFileIds string `form:"userFileIds"` // 多个下载文件的用户文件标识符，以逗号分割
}

// FilePreviewReq 文件在线预览请求
type FilePreviewReq struct {
	UserFileId     string `form:"userFileId"`     // 用户文件标识符
	IsMin          bool   `form:"isMin"`          // 是否是以最低质量预览
	ShareBatchNum  string `form:"shareBatchNum"`  // 未使用
	ExtractionCode string `form:"extractionCode"` // 未使用
}

// ----------------------------------------------------------------

// DeleteFileReq 文件批量删除请求
type DeleteFileReq struct {
	UserFileId string `json:"userFileId"` // 要删除文件的用户文件标识符
}

// DeleteFileInBatchReq 文件批量删除请求
type DeleteFileInBatchReq struct {
	UserFileIds string `json:"userFileIds"` // 要批量删除的文件的用户文件标识符，以逗号隔开
}

// ----------------------------------------------------------------

// CreateFileReq 创建文件请求API
type CreateFileReq struct {
	FilePath   string `json:"filePath"`   // 文件路径
	FileName   string `json:"fileName"`   // 文件名
	ExtendName string `json:"extendName"` // 扩展名
}

// CreateFolderReq 创建文件夹请求API
type CreateFolderReq struct {
	FolderName string `json:"fileName"` // 文件路径
	FolderPath string `json:"filePath"` // 文件名
}

// ----------------------------------------------------------------

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
	FileName string `form:"filename"`
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
