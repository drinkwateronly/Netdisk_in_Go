package utils

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type H struct {
	Success bool        `json:"success"`
	Msg     string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Resp(w http.ResponseWriter, code int, success bool, data interface{}, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	h := H{
		Success: success,
		Data:    data,
		Msg:     msg,
	}
	ret, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
	}
	_, err = w.Write(ret)
	if err != nil {
		panic(err)
	}
}

func RespList(w http.ResponseWriter, dataList interface{}, total int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := gin.H{
		"success":  true,
		"dataList": dataList,
		"total":    total,
		"message":  msg,
	}
	ret, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(ret)
}

func RespFail(w http.ResponseWriter, msg string) {
	Resp(w, http.StatusBadRequest, false, nil, msg)
}

func RespOK(w http.ResponseWriter, data interface{}, msg string) {
	Resp(w, http.StatusOK, true, data, msg)
}

//func RespOKList(w http.ResponseWriter, data interface{}, msg string) {
//	RespList(w, 0, data, msg)
//}
