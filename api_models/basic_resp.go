package api_models

type RespData struct {
	Code    int         `json:"code" binding:"required"`
	Success bool        `json:"success" binding:"required"`
	Data    interface{} `json:"data" binding:"required"`
	Message string      `json:"message" binding:"required"`
}

type RespDataList struct {
	Code     int         `json:"code" binding:"required"`
	Success  bool        `json:"success" binding:"required"`
	DataList interface{} `json:"dataList" binding:"required"`
	Total    int         `json:"total" binding:"required"`
	Message  string      `json:"message" binding:"required"`
}
