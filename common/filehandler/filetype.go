package filehandler

import "strings"

const (
	OTHER = iota
	IMAGE
	EDITABLE
	VIDEO
	AUDIO
	DIRECTORY
	ROOT
)

var fileTypeMap = map[string]uint8{
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

// 根据文件拓展名获取文件类型
func getFileType(extend string) uint8 {
	// extend在map中的key不存在时默认为OTHER类型，即0值
	return fileTypeMap[strings.ToLower(extend)]
}
