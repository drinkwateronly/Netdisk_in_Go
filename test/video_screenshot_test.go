package test

import (
	"io"
	"netdisk_in_go/utils"
	"os"
	"testing"
)

// ffmpeg，下载并假如环境变量即可
// https://zhuanlan.zhihu.com/p/118362010
// https://github.com/BtbN/FFmpeg-Builds/releases
func TestScreenShot(t *testing.T) {
	reader, err := utils.GetFrameFromVideo("E:\\go\\netdisk_in_go\\repository\\upload_file\\4545e03b-dd30-41f1-b034-fdd2c708223e", 5)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile("./image/frame.jpeg", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0777)
	defer file.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(file, reader)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVideoClip(t *testing.T) {
	err := utils.SavePreviewFromVideo("E:\\go\\netdisk_in_go\\repository\\upload_file\\4545e03b-dd30-41f1-b034-fdd2c708223e", 5)
	if err != nil {
		t.Fatal(err)
	}
	//file, err := os.OpenFile("./video/clip.mp4", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//_, err = io.Copy(file, reader)
	//if err != nil {
	//	t.Fatal(err)
	//}
}
