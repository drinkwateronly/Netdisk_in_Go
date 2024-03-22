package test

import (
	"io"
	"netdisk_in_go/common"
	"os"
	"testing"
)

func TestImageCompress(t *testing.T) {
	file, err := os.Open("./image/IMG_7140.png")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	newFile, err := common.CompressImage(file, 100, 50, "")
	//newFile := common.CompressImageResource(file)
	if err != nil {
		panic(err)

	}

	saveFile, _ := os.Create("./image/compressPng.jpg")
	defer saveFile.Close()
	io.Copy(saveFile, newFile)
}
