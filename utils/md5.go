package utils

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5Encode(data string) string {

	h := md5.New()
	h.Write([]byte(data))
	tmpStr := h.Sum([]byte(nil))
	return hex.EncodeToString(tmpStr)
}

func Md5EncodeByte(data []byte) string {
	h := md5.New()
	h.Write(data)
	tmpStr := h.Sum([]byte(nil))
	return hex.EncodeToString(tmpStr)
}
