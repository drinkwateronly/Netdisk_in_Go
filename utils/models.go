package utils

import (
	ApiModels "netdisk_in_go/api_models"
	"strings"
)

type ProcessedFileInfo struct {
	// 假设文件为路径为 "/456/789/0.txt"
	FullPath   string // "0"
	AbsPath    string // /456/789，数据库存放的文件目录
	FileName   string // "txt
	ExtendName string
	FileType   uint8
}

// GetFileInfoFromReq 从文件上传请求参数ApiModels.FileUploadReqAPI中处理出用于文件上传的必须参数，以ProcessedFileInfo保存
func GetFileInfoFromReq(req ApiModels.FileUploadReqAPI) ProcessedFileInfo {
	var fileInfo ProcessedFileInfo

	// 文件名与文件拓展名
	split := strings.Split(req.FileFullName, ".")
	if len(split) == 1 { // 没有文件拓展名
		fileInfo.ExtendName = ""
		fileInfo.FileName = req.FileFullName
	} else {
		fileInfo.ExtendName = split[len(split)-1]
		fileInfo.FileName = req.FileFullName[0 : len(req.FileFullName)-len(fileInfo.ExtendName)-1] // 去掉文件全名右侧的拓展名
	}

	// 文件拓展名映射为文件类型
	fileInfo.FileType = FileTypeId[fileInfo.ExtendName]
	if req.IsDir == 1 {
		fileInfo.FileType = 6 // 文件夹
	} else if fileInfo.FileType == 0 {
		fileInfo.FileType = 5 // 其他
	}

	// 包括文件名的完整路径fileFullPath
	var fileFullPath string
	if req.FilePath == "/" {
		fileFullPath = "/" + req.RelativePath
	} else {
		fileFullPath = req.FilePath + "/" + req.RelativePath
	}
	if len(fileFullPath) == len(req.FileFullName)+1 { // 即多出了一个/
		fileInfo.AbsPath = "/" // 不去掉最后的"/"
	} else {
		fileInfo.AbsPath = fileFullPath[:len(fileFullPath)-len(req.FileFullName)-1] // 去掉最后的"/"
	}

	return fileInfo
}
