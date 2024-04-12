package api

// FileShareReq 文件分享请求API
type FileShareReq struct {
	EndTime     string `form:"endTime"`     // 分享结束时长
	Remark      string `form:"remark"`      // 未使用
	ShareType   uint8  `form:"shareType"`   // 分享类型，有验证码时为1
	UserFileIds string `form:"userFileIds"` // 分享的用户文件ids
}

// FileShareResp 文件分享响应api
type FileShareResp struct {
	ShareBatchId   string `json:"shareBatchNum"`  // 分享批次
	ExtractionCode string `json:"extractionCode"` // 提取码
}

// CheckShareReq 检查分享过期时间或分享类型请求api
type CheckShareReq struct {
	ShareBatchId string `form:"shareBatchNum"`
}

// CheckShareTypeResp 检查分享类型api
type CheckShareTypeResp struct {
	ShareType uint8 `json:"shareType"`
}

// CheckExtractionCodeReq 检查分享验证码请求api
type CheckExtractionCodeReq struct {
	ShareBatchId   string `form:"shareBatchNum"`
	ExtractionCode string `form:"extractionCode"`
}

type SaveShareReq struct {
	FilePath      string `form:"filePath"`
	UserFileIds   string `form:"userFileIds"`
	ShareBatchNum string `form:"shareBatchNum"`
}

type GetShareFileListReq struct {
	ShareBatchId  string `form:"shareBatchNum"` // 分享批次id
	ShareFilePath string `form:"shareFilePath"` // 分享批次内路径
}

type GetShareFileListResp struct {
	UserFileId    string `json:"userFileId"`
	ShareBatchId  string `json:"shareBatchNum"`
	ShareFilePath string `json:"shareFilePath"`
	FileName      string `json:"fileName"`
	ExtendName    string `json:"extendName"`
	FileSize      uint64 `json:"fileSize"`
	FileType      uint8  `json:"fileType"`
	IsDir         uint8  `json:"isDir"`
}

type GetShareListReq struct {
	ShareBatchId  string `form:"shareBatchNum"` // 分享批次id
	ShareFilePath string `form:"shareFilePath"`
	CurrentPage   uint   `form:"currentPage"` // 分享批次内路径
	PageCount     uint   `form:"pageCount"`
}

type GetMyShareListResp struct {
	UserFileId    string `json:"userFileId"`
	UserId        string `json:"userId"`
	ShareBatchId  string `json:"shareBatchNum" gorm:"column:share_batch_id"`
	ShareFilePath string `json:"shareFilePath"`
	ShareType     uint8  `json:"shareType"`
	ParentId      string `json:"parentId"`
	FileName      string `json:"fileName"`
	ExtendName    string `json:"extendName"`
	FileType      uint8  `json:"fileType"`
	IsDir         uint8  `json:"isDir"`
	FileSize      uint64 `json:"fileSize"`
	ModifyTime    string `json:"modifyTime"`
	UploadTime    string `json:"uploadTime"`
	EndTime       string `json:"endTime" gorm:"column:expire_time"`
}
