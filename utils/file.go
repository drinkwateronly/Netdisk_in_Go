package utils

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	IMAGE = iota + 1
	EDITABLE
	VIDEO
	AUDIO
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

// GetFrameFromVideo 读取路径的视频文件，并截取frameIndex对应帧
func GetFrameFromVideo(videoFilePath string, frameIndex int) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(videoFilePath).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameIndex)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	return buf, err
}

// GetClipFromVideo 读取路径的视频文件，并截取一段视频，已弃用
func GetClipFromVideo(videoFilePath string) (io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(videoFilePath).
		Filter("trim", ffmpeg.Args{fmt.Sprintf("start=0:end=120")}). // 截取两分钟的视频
		Filter("scale", ffmpeg.Args{"640:480"}).                     // 将分辨率变小
		Output(videoFilePath + "-pv").
		Run()
	//ffmpeg.KwArgs{"ss": 120, "t": 120, "b:v": "512k"}
	return buf, err
}

// SavePreviewFromVideo 读取videoFilePath路径的视频文件，并截取frameIndex对应帧作为视频文件的preview
func SavePreviewFromVideo(videoFilePath string, frameIndex int) error {
	err := ffmpeg.Input(videoFilePath).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameIndex)}).
		Output(videoFilePath+"-pv", ffmpeg.KwArgs{"vframes": 1, "format": "mp4", "vcodec": "libx264"}).
		Run()
	return err
}

func SavePreviewFromImage(imageFilePath, imageType string) error {
	// 读取图片文件
	file, err := os.OpenFile(imageFilePath, os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	// 压缩图像
	compressImage, _ := CompressImage(file, 50, 50, imageType)
	//if err != nil {
	//	return err
	//}
	// 存放图片的preview文件
	create, err := os.Create(imageFilePath + "-pv")
	if err != nil {
		return err
	}
	_, err = io.Copy(create, compressImage)
	if err != nil {
		return err
	}
	return err
}

// CompressImage 将图像按类型进行压缩，并指定最大的高度maxHeight，压缩过程出现任何error，则会返回未压缩的图像
// imageType 有jpg/jpeg、gif、png 3种
// compressQuality 仅用于jpg图像压缩
func CompressImage(file io.Reader, maxHeight uint, compressQuality int, imageType string) (io.Reader, error) {
	//var maxHeight uint = 100 // 设置压缩图像的最高高度
	//compressQuality := 50    // 压缩图像的质量，越小越差
	imageType = strings.ToLower(imageType)
	img, _, err := image.Decode(file)
	if err != nil {
		// 并非图像文件
		return file, err
	}

	//// 备用方案代码获取图像尺寸，并等比缩放，但没有效果
	//// 获取图像尺寸
	//width := img.Bounds().Max.X - img.Bounds().Min.X
	//height := img.Bounds().Max.Y - img.Bounds().Min.Y
	//// 修改图像尺寸
	//compressImage := image.NewRGBA(image.Rect(0, 0, width*maxHeight/height, maxHeight))
	//draw.Draw(compressImage, compressImage.Bounds(), img, image.Point{}, draw.Src)

	resizedImg := resize.Resize(0, maxHeight, img, resize.Lanczos3)
	buf := bytes.Buffer{}

	if imageType == "jpg" || imageType == "jpeg" {
		// 如果是jpg
		err = jpeg.Encode(&buf, resizedImg, &jpeg.Options{Quality: compressQuality})
	} else if imageType == "gif" {
		// 如果是gif
		err = gif.Encode(&buf, resizedImg, &gif.Options{})
	} else {
		// 剩下的用png
		err = png.Encode(&buf, resizedImg)
	}

	if err != nil {
		return file, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}
