package api_models

type CheckShareTypeResp struct {
	ShareType uint8 `json:"shareType"`
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
	ShareFilePath string `form:"shareFilePath"` // 分享批次id
	CurrentPage   uint   `form:"currentPage"`   // 分享批次内路径
	PageCount     uint   `form:"pageCount"`
}

type GetShareListResp struct {
	UserFileId   string `json:"userFileId"`
	UserId       string `json:"userId"`
	FilePath     string `json:"filePath"`
	ParentId     string `json:"parentId"`
	FileName     string `json:"fileName"`
	ExtendName   string `json:"extendName"`
	FileType     uint8  `json:"fileType"`
	IsDir        uint8  `json:"isDir"`
	FileSize     uint64 `json:"fileSize"`
	ModifyTime   string `json:"modifyTime"`
	UploadTime   string `json:"uploadTime"`
	EndTime      string `json:"endTime" gorm:"column:expire_time"`
	ShareBatchId string `json:"shareBatchNum" gorm:"column:share_batch_id"`
}
