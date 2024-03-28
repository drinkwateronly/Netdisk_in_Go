package api

type RecoveryListResp struct {
	UserFileId    string `json:"userFileId"`
	UserId        string `json:"userId"`
	DeleteBatchId string `json:"deleteBatchNum"`
	FilePath      string `json:"filePath"`
	FileName      string `json:"fileName"`
	FileType      uint8  `json:"fileType"`
	ExtendName    string `json:"extendName"`
	IsDir         uint8  `json:"isDir"`
	FileSize      uint64 `json:"fileSize"`
	DeleteTime    string `json:"deleteTime"`
	UploadTime    string `json:"uploadTime"`
}

// DelRecoveryReq 删除回收站文件请求
type DelRecoveryReq struct {
	UserFileId string `json:"userFileId"`
}

// DelRecoveryInBatchReq 批量删除回收站文件请求
type DelRecoveryInBatchReq struct {
	UserFileIds string `json:"userFileIds"`
}

type RecoverFileReq struct {
	DeleteBatchNum string `json:"deleteBatchNum"` // 删除的批次
	FilePath       string `json:"filePath"`       // 恢复的路径
}
