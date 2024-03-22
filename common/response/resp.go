package response

import (
	"encoding/json"
	"net/http"
)

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

func RespOK(w http.ResponseWriter, code int, success bool, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	ret, _ := json.Marshal(RespData{
		Code:    code,
		Success: success,
		Data:    data,
		Message: msg,
	})
	_, err := w.Write(ret)
	if err != nil {
		panic(err)
	}
}

func RespOKFail(w http.ResponseWriter, code int, data interface{}, msg string) {
	RespOK(w, code, false, data, msg)
}

func RespOKSuccess(w http.ResponseWriter, code int, data interface{}, msg string) {
	RespOK(w, code, true, data, msg)
}

func RespBadReq(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	ret, _ := json.Marshal(RespData{
		Code:    99999,
		Success: false,
		Data:    nil,
		Message: msg,
	})
	_, err := w.Write(ret)
	if err != nil {
		panic(err)
	}
}

func RespOkWithDataList(w http.ResponseWriter, code int, dataList interface{}, total int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var ret []byte
	if total == 0 {
		// 此时切片为空，前端返回时只会收到nil，导致无法对nil遍历，因此需要返回一个数组
		ret, _ = json.Marshal(RespDataList{
			Code:     code,
			Success:  true,
			DataList: [0]interface{}{}, // [0]的原因是需要其在没有数据的时候仍然返回一个切片
			Total:    total,
			Message:  msg,
		})
	} else {
		ret, _ = json.Marshal(RespDataList{
			Code:     code,
			Success:  true,
			DataList: dataList,
			Total:    total,
			Message:  msg,
		})
	}
	_, err := w.Write(ret)
	if err != nil {
		panic(err)
	}
}
