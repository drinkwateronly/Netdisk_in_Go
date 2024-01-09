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
func TestScreenshot(t *testing.T) {
	reader, err := utils.GetFrameFromVideo("E:\\go\\netdisk_in_go\\repository\\upload_file\\4545e03b-dd30-41f1-b034-fdd2c708223e", 5)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Create("./image/frame.jpeg")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(file, reader)
	if err != nil {
		t.Fatal(err)
	}

}
