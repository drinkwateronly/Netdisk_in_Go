package api

type RecoveryListRespAPI struct {
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

type DelRecoveryReqAPI struct {
	UserFileId string `json:"userFileId"`
}

type DelRecoveryFilesInBatchReq struct {
	UserFileIds string `json:"userFileIds"`
}
