package test

import (
	"io"
	"netdisk_in_go/utils"
	"os"
	"testing"
)

func TestImageCompress(t *testing.T) {
	file, err := os.Open("./image/IMG_7140.png")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	newFile, err := utils.CompressImage(file, 100, 50, "")
	//newFile := utils.CompressImageResource(file)
	if err != nil {
		panic(err)

	}

	saveFile, _ := os.Create("./image/compressPng.jpg")
	defer saveFile.Close()
	io.Copy(saveFile, newFile)
}
