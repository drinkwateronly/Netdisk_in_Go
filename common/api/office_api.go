package api

type PrepareOnlyOfficeReq struct {
	UserFileId string `json:"userFileId"` // 用户文件id
}

type OfficeFilePreviewReq struct {
	UserFileId string `form:"userFileId"` // 用户文件id
	Cookie     string `form:"token"`      // cookie
}

type OfficeErrorResp struct {
	Error int `json:"error"`
}

type OfficeCallbackReq struct {
	Actions       interface{} `json:"actions"`        // actions:[map[type:1 userid:001]]
	ChangeHistory interface{} `json:"changeshistory"` //
	ChangesURL    string      `json:"changesurl"`
	FileType      string      `json:"filetype"`
	ForceSaveType int         `json:"forcesavetype"`
	History       interface{} `json:"history"`
	Key           string      `json:"key"`
	Status        int         `json:"status"`
	Url           string      `json:"url"`
	UserData      string      `json:"userdata"`
	Users         []string    `json:"users"`
}
