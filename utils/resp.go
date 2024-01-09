package utils

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

type H struct {
	Code    int         `json:"code"`
	Success bool        `json:"success"`
	Msg     string      `json:"message"`
	Data    interface{} `json:"data"`
}

func RespOK(w http.ResponseWriter, code int, success bool, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := H{
		Code:    code,
		Success: success,
		Data:    data,
		Msg:     msg,
	}
	ret, _ := json.Marshal(h)
	_, err := w.Write(ret)
	if err != nil {
		panic(err)
	}
}

func RespBadReq(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	h := H{
		Code:    999999,
		Success: false,
		Data:    nil,
		Msg:     msg,
	}
	ret, _ := json.Marshal(h)
	_, err := w.Write(ret)
	if err != nil {
		panic(err)
	}
}

func RespOkWithDataList(w http.ResponseWriter, code int, dataList interface{}, total int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := gin.H{
		"code":     code,
		"success":  true,
		"dataList": dataList,
		"total":    total,
		"message":  msg,
	}
	ret, _ := json.Marshal(h)
	_, err := w.Write(ret)
	if err != nil {
		panic(err)
	}
}

//func RespOKList(w http.ResponseWriter, data interface{}, msg string) {
//	RespList(w, 0, data, msg)
//}
