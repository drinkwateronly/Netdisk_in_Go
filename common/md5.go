package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// 在线计算文件md5，用于测试：https://www.nuomiphp.com/filemd5.html

func GetFileMd5(file *os.File) (string, error) {
	md5 := md5.New()
	_, err := io.Copy(md5, file)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	md5Str := hex.EncodeToString(md5.Sum(nil))
	return md5Str, nil
}

func GetFileMD5FromPath(path string) (string, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0777)
	if err != nil {
		return "", err
	}
	return GetFileMd5(file)
}

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
