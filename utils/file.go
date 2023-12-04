package utils

import (
	"os"
	"strconv"
)

func MergeChunkToFile(fileName string, totalChunks int) error {
	// 创建一个大文集
	myFileCopy, err := os.OpenFile("./repository/upload_file/"+fileName, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	// 记录目前文件的末尾
	var fileEnd int64
	// 循环所有的chunk
	for i := 1; i <= totalChunks; i++ {
		// 查看chunk的文件信息
		chuckFilePath := "./repository/chunk_file/" + fileName + "-" + strconv.Itoa(i) + ".chunk"
		fileInfo, err := os.Stat(chuckFilePath)
		if err != nil {
			return err
		}
		b := make([]byte, fileInfo.Size())

		// 打开chuck文件
		f, err := os.OpenFile(chuckFilePath, os.O_RDONLY, 0777)
		if err != nil {
			MyLog.Println(err)
			return err
		}
		f.Read(b)
		myFileCopy.Seek(fileEnd, 0)
		myFileCopy.Write(b)
		f.Close()
		fileEnd += fileInfo.Size()
	}
	myFileCopy.Close()
	return nil
}
