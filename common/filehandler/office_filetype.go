package filehandler

import "strings"

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
