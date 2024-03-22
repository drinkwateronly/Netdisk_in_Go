package utils

import (
	"encoding/json"
	"net/http"
	"netdisk_in_go/models/api_models"
)

func RespOK(w http.ResponseWriter, code int, success bool, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	ret, _ := json.Marshal(api_models.RespData{
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

func RespBadReq(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	ret, _ := json.Marshal(api_models.RespData{
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
		ret, _ = json.Marshal(api_models.RespDataList{
			Code:     code,
			Success:  true,
			DataList: [0]interface{}{}, // [0]的原因是需要其在没有数据的时候仍然返回一个切片
			Total:    total,
			Message:  msg,
		})
	} else {
		ret, _ = json.Marshal(api_models.RespDataList{
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
