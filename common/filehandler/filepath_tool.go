package filehandler

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// ConCatFileFullPath 根据记录中的文件路径（其父文件夹的绝对路径）与文件名拼接成文件完整路径（绝对路径）
// example：
// - "/123/456" + "789" ->  "/123/456/789"
// - "/123/456" + "789.txt" ->  "/123/456/789.txt"
// - "/" + "789.txt" ->  "/789.txt"
func ConCatFileFullPath(filePath, fileName string) string {
	if filePath == "/" {
		return "/" + fileName
	}
	return filePath + "/" + fileName
}

// SplitAbsPath 从绝对路径中分割出文件/文件夹名称 及其 父文件夹绝对路径，返回路径是否合法错误
func SplitAbsPath(absPath string) (parentAbsPath, fileName string, err error) {
	index := strings.LastIndex(absPath, "/")
	// 地址不合法：（1）"/"不存在 （2）"/"在最后一位，即该绝对路径不包含文件/文件夹
	if index == -1 || index == len(absPath)-1 {
		return "", "", errors.New("invalid path: " + absPath)
	}
	// 文件或文件夹存在为根目录，例如"/123"
	if index == 0 {
		return "/", absPath[index+1:], nil
	}
	return absPath[:index], absPath[index+1:], nil
}

// RenameConflictFile 从绝对路径中分割出文件/文件夹名称 及其 父文件夹绝对路径，返回路径是否合法错误
func RenameConflictFile(fileName string) string {
	pattern := "\\(\\d+\\)$" // 正则表达式匹配 以"(数字)"结尾的字符串
	re := regexp.MustCompile(pattern)
	match := re.FindString(fileName)
	if match == "" || match == fileName { // 没匹配上，或文件名就是"(数字)"
		return fileName + "(1)"
	}
	num, _ := strconv.Atoi(match[1 : len(match)-1])
	return fileName[:len(fileName)-len(match)] + "(" + strconv.Itoa(num+1) + ")"
}
