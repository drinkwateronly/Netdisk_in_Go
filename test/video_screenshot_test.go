package test

import (
	"bytes"
	"fmt"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"io"
	"os"
	"testing"
)

func TestScreenshot(t *testing.T) {
	reader := exampleReadFrameAsJpeg("E:\\go\\netdisk_in_go\\repository\\upload_file\\4545e03b-dd30-41f1-b034-fdd2c708223e", 5)
	file, err := os.Create("./out1.jpeg")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(file, reader)
	if err != nil {
		t.Fatal(err)
	}

}

func exampleReadFrameAsJpeg(inFileName string, frameNum int) io.Reader {
	buf := bytes.NewBuffer(nil)

	err := ffmpeg.Input(inFileName).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		panic(err)
	}
	return buf
}
