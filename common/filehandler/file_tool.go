package filehandler

import (
	"bytes"
	"fmt"
	"github.com/nfnt/resize"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"netdisk_in_go/common/api"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func IsChunkExist(filename string, currentChunkSize uint) bool {
	fileInfo, err := os.Stat(filename)
	return !os.IsNotExist(err) && uint(fileInfo.Size()) == currentChunkSize
}

//func DeleteAllChunks(chuckName string, totalChunks int) {
//	for i := 1; i <= totalChunks; i++ {
//		// 检查文件是否存在
//		chunkFilePath := "./repository/chunk_file/" + chuckName + "-" + strconv.Itoa(i) + ".chunk"
//		// 由于分片是小编号开始，所以一旦某个分片不存在，其后续编号的分片也不会存在，删除所有分片即完成
//		if IsChunkExist(chunkFilePath) {
//			return
//		}
//		os.Remove(chunkFilePath)
//	}
//}

func MergeChunksToFile(chuckName, fileName string, totalChunks uint) error {
	// 创建一个用于存放大文件的新文件
	completeFile, err := os.OpenFile("./repository/upload_file/"+fileName, os.O_CREATE|os.O_RDWR, 0777)
	defer completeFile.Close()
	if err != nil {
		return err
	}
	// 记录目前文件的末尾
	var fileEnd int64
	// 循环所有的chunk
	//fmt.Fprint(gin.DefaultWriter, "totalChunks", totalChunks) // 打印一下
	for i := uint(1); i <= totalChunks; i++ {
		// 获取分片文件大小
		chunkFilePath := "./repository/chunk_file/" + chuckName + "-" + strconv.FormatInt(int64(i), 10) + ".chunk"
		fileInfo, err := os.Stat(chunkFilePath)
		if err != nil {
			return err
		}
		b := make([]byte, fileInfo.Size())
		// 读取分片
		chunkFile, err := os.OpenFile(chunkFilePath, os.O_RDONLY, 0777)
		if err != nil {
			return err
		}
		// 读取分片文件
		_, err = chunkFile.Read(b)
		if err != nil {
			return err
		}
		// 往大文件尾部填充分片
		_, err = completeFile.Seek(fileEnd, 0)
		if err != nil {
			return err
		}
		_, err = completeFile.Write(b)
		if err != nil {
			return err
		}
		// 可以直接关闭分片
		err = chunkFile.Close()
		if err != nil {
			return err
		}
		// 更新大文件尾部指针
		fileEnd += fileInfo.Size()
	}
	return nil
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

type ProcessedFileInfo struct {
	// 假设文件为路径为 "/456/789/0.txt"
	AbsPath    string // "/456/789/0.txt"
	FilePath   string // "/456/789"
	FileName   string // "0.txt"
	ExtendName string // "txt"
	FileType   uint8  //
}

// GetFileInfoFromReq 从文件上传请求参数api.FileUploadReqAPI中处理出用于文件上传的必须参数，以ProcessedFileInfo保存
func GetFileInfoFromReq(req api.FileUploadReqAPI) ProcessedFileInfo {
	var fileInfo ProcessedFileInfo
	// 文件名与文件拓展名
	ext := filepath.Ext(req.FileName) // 拓展名前带'.'
	fileInfo.FileName = req.FileName[:len(req.FileName)-len(ext)]
	if len(ext) != 0 {
		ext = ext[1:] // 去掉 '.'
	}
	fileInfo.ExtendName = ext

	// 文件拓展名映射为文件类型
	fileInfo.FileType = getFileType(ext)
	if req.IsDir == 1 {
		fileInfo.FileType = DIRECTORY // 文件夹
	}

	// 拼接出文件绝对路径，包括文件名，例如"/456/123.txt"
	absPath := ConCatFileFullPath(req.FilePath, req.RelativePath)

	if absPath == "/"+req.FileName { // 或者len(absPath) == len(req.FileName)+1
		// 上传文件存放到网盘根目录，例如absPath = "/123.txt"
		fileInfo.FilePath = "/"
	} else {
		// 上传文件不存放到网盘根目录，例如absPath = "/456/123.txt"
		fileInfo.FilePath = absPath[:len(absPath)-len(req.FileName)-1] // 去掉最后的"/123.txt"
	}
	return fileInfo
}
