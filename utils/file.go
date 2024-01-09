package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	IMAGE = iota + 1
	EDITABLE
	AUDIO
	VIDEO
	DICTIONARY
	OTHER
)

var FileTypeId = map[string]int{
	"bmp":      IMAGE,
	"jpg":      IMAGE,
	"png":      IMAGE,
	"tif":      IMAGE,
	"gif":      IMAGE,
	"jpeg":     IMAGE,
	"doc":      EDITABLE,
	"docx":     EDITABLE,
	"docm":     EDITABLE,
	"dot":      EDITABLE,
	"dotx":     EDITABLE,
	"dotm":     EDITABLE,
	"odt":      EDITABLE,
	"fodt":     EDITABLE,
	"ott":      EDITABLE,
	"rtf":      EDITABLE,
	"txt":      EDITABLE,
	"html":     EDITABLE,
	"htm":      EDITABLE,
	"mht":      EDITABLE,
	"xml":      EDITABLE,
	"pdf":      EDITABLE,
	"djvu":     EDITABLE,
	"fb2":      EDITABLE,
	"epub":     EDITABLE,
	"xps":      EDITABLE,
	"xls":      EDITABLE,
	"xlsx":     EDITABLE,
	"xlsm":     EDITABLE,
	"xlt":      EDITABLE,
	"xltx":     EDITABLE,
	"xltm":     EDITABLE,
	"ods":      EDITABLE,
	"fods":     EDITABLE,
	"ots":      EDITABLE,
	"csv":      EDITABLE,
	"pps":      EDITABLE,
	"ppsx":     EDITABLE,
	"ppsm":     EDITABLE,
	"ppt":      EDITABLE,
	"pptx":     EDITABLE,
	"pptm":     EDITABLE,
	"pot":      EDITABLE,
	"potx":     EDITABLE,
	"potm":     EDITABLE,
	"odp":      EDITABLE,
	"fodp":     EDITABLE,
	"otp":      EDITABLE,
	"hlp":      EDITABLE,
	"wps":      EDITABLE,
	"java":     EDITABLE,
	"json":     EDITABLE,
	"css":      EDITABLE,
	"go":       EDITABLE,
	"py":       EDITABLE,
	"c":        EDITABLE,
	"cpp":      EDITABLE,
	"markdown": EDITABLE,
	"md":       EDITABLE,
	"avi":      VIDEO,
	"mp4":      VIDEO,
	"mpg":      VIDEO,
	"mov":      VIDEO,
	"swf":      VIDEO,
	"wav":      AUDIO,
	"aif":      AUDIO,
	"au":       AUDIO,
	"mp3":      AUDIO,
	"ram":      AUDIO,
	"wma":      AUDIO,
	"mmf":      AUDIO,
	"amr":      AUDIO,
	"aac":      AUDIO,
	"flac":     AUDIO,
}

const (
	WORD = iota + 1
	CELL
	SLIDE
)

var FileOfficeTypeId = map[string]int{
	"djvu":  WORD,
	"doc":   WORD,
	"docm":  WORD,
	"docx":  WORD,
	"docxf": WORD,
	"dot":   WORD,
	"dotm":  WORD,
	"dotx":  WORD,
	"epub":  WORD,
	"fb2":   WORD,
	"fodt":  WORD,
	"htm":   WORD,
	"html":  WORD,
	"mht":   WORD,
	"mhtml": WORD,
	"odt":   WORD,
	"oform": WORD,
	"ott":   WORD,
	"oxps":  WORD,
	"pdf":   WORD,
	"rtf":   WORD,
	"stw":   WORD,
	"sxw":   WORD,
	"txt":   WORD,
	"wps":   WORD,
	"wpt":   WORD,
	"xps":   WORD,
	//"xml":   WORD,

	"csv":  CELL,
	"et":   CELL,
	"ett":  CELL,
	"fods": CELL,
	"ods":  CELL,
	"ots":  CELL,
	"sxc":  CELL,
	"xls":  CELL,
	"xlsb": CELL,
	"xlsm": CELL,
	"xlsx": CELL,
	"xlt":  CELL,
	"xltm": CELL,
	"xltx": CELL,
	"xml":  CELL,

	"dps":  SLIDE,
	"dpt":  SLIDE,
	"fodp": SLIDE,
	"odp":  SLIDE,
	"otp":  SLIDE,
	"pot":  SLIDE,
	"potm": SLIDE,
	"potx": SLIDE,
	"pps":  SLIDE,
	"ppsm": SLIDE,
	"ppsx": SLIDE,
	"ppt":  SLIDE,
	"pptm": SLIDE,
	"pptx": SLIDE,
	"sxi":  SLIDE,
}

func GetOfficeDocumentType(extendName string) (string, bool) {
	extendName = strings.ToLower(extendName)
	switch FileOfficeTypeId[extendName] {
	case CELL:
		return "Cell", true
	case WORD:
		return "Word", true
	case SLIDE:
		return "Slide", true
	}
	return "", false
}

func MergeChunkToFile(chuckName, fileName string, totalChunks int) (string, error) {
	// 创建一个大文集
	myFileCopy, err := os.OpenFile("./repository/upload_file/"+fileName, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return "", err
	}
	// 记录目前文件的末尾
	var fileEnd int64
	// 循环所有的chunk
	fmt.Fprint(gin.DefaultWriter, "totalChunks", totalChunks) // 打印一下

	for i := 1; i <= totalChunks; i++ {

		// 查看chunk的文件信息
		chuckFilePath := "./repository/chunk_file/" + chuckName + "-" + strconv.Itoa(i) + ".chunk"
		fileInfo, err := os.Stat(chuckFilePath)
		if err != nil {
			return "", err
		}
		b := make([]byte, fileInfo.Size())

		// 打开chuck文件
		f, err := os.OpenFile(chuckFilePath, os.O_RDONLY, 0777)
		if err != nil {
			MyLog.Println(err)
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
	fileMD5 := Md5EncodeByte(fileByte)

	return fileMD5, nil
}
