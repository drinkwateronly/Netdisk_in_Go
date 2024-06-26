package test

import (
	"fmt"
	"io"
	"netdisk_in_go/common"
	"netdisk_in_go/common/filehandler"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func MergeChunkToFile(fileName string, totalChunks int) (string, error) {
	// 创建一个大文集
	myFileCopy, err := os.OpenFile("../repository/upload_file/"+fileName, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return "", err
	}
	// 记录目前文件的末尾
	var fileEnd int64
	// 循环所有的chunk
	for i := 1; i <= totalChunks; i++ {
		// 查看chunk的文件信息
		chuckFilePath := "../repository/chunk_file/" + fileName + "-" + strconv.Itoa(i) + ".chunk"
		fileInfo, err := os.Stat(chuckFilePath)
		if err != nil {
			return "", err
		}
		b := make([]byte, fileInfo.Size())

		// 打开chuck文件
		f, err := os.OpenFile(chuckFilePath, os.O_RDONLY, 0777)
		if err != nil {
			return "", err
		}
		f.Read(b)
		myFileCopy.Seek(fileEnd, 0)
		myFileCopy.Write(b)
		f.Close()
		fileEnd += fileInfo.Size()
	}
	myFileCopy.Close()
	fileByte, _ := io.ReadAll(myFileCopy)
	fileMD5 := common.Md5EncodeByte(fileByte)
	return fileMD5, nil
}

func TestFileMerge(t *testing.T) {
	mergeFileMD5, err := MergeChunkToFile("d45eda1235e0dfb8546c0427ac1b606f", 170)
	fmt.Printf(mergeFileMD5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileHash(t *testing.T) {
	filePath1 := "E:\\go\\netdisk_in_go\\repository\\upload_file\\d45eda1235e0dfb8546c0427ac1b606f"
	filePath2 := "E:\\go\\netdisk_in_go\\repository\\upload_file\\QQ9.9.6.19189_x64.exe"
	file1, _ := os.OpenFile(filePath1, os.O_RDONLY, 0777)
	file2, _ := os.OpenFile(filePath2, os.O_RDONLY, 0777)
	fileByte1, _ := io.ReadAll(file1)
	fileByte2, _ := io.ReadAll(file2)
	hash1 := common.Md5EncodeByte(fileByte1)
	hash2 := common.Md5EncodeByte(fileByte2)
	fmt.Println(hash1)
	if hash1 != hash2 {
		t.Fatal(hash1, hash2)
	}
}

func TestOpenFile(t *testing.T) {
	file, err := os.OpenFile("../repository/upload_file/13df9051-a743-4870-bc6f-4b07939d48bf", os.O_RDONLY, 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
}

func TestMD5(t *testing.T) {
	file, err := os.OpenFile("../repository/upload_file/13df9051-a743-4870-bc6f-4b07939d48bf", os.O_RDONLY, 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	md5, err := common.GetFileMd5(file)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(md5)
}

func TestFileIsExist(t *testing.T) {
	info, err := os.Stat("./file_test.go")
	if os.IsNotExist(err) {
		t.Fatal("文件不存在，但实际存在")
	} else {
		fmt.Println(info.Size())
	}

}

func TestPathSplit(t *testing.T) {
	fmt.Println(filehandler.SplitAbsPath("/"))
	fmt.Println(filehandler.SplitAbsPath("/abc/123"))
	fmt.Println(filehandler.SplitAbsPath("/abc/ass.123"))
}

func TestRenameConflictFile(t *testing.T) {
	fmt.Println(filehandler.RenameConflictFile("123"))
	fmt.Println(filehandler.RenameConflictFile("(1)"))
	fmt.Println(filehandler.RenameConflictFile("a(1)"))

	fmt.Println("/abc/123"[len("/abc"):])
}

func TestGetFileExt(t *testing.T) {
	fileName := "1.txt"
	fmt.Println(filepath.Ext(fileName))
	fileName = "/232/1.txt"
	fmt.Println(filepath.Ext(fileName))
	fileName = "/232/1"
	fmt.Println(filepath.Ext(fileName))
	fileName = "/232/1.7z.001"
	fmt.Println(filepath.Ext(fileName))
}
