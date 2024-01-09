package test

import (
	"io"
	"netdisk_in_go/utils"
	"os"
	"testing"
)

func TestImageCompress(t *testing.T) {
	file, err := os.Open("./image/origin.jpg")
	defer file.Close()
	if err != nil {
		panic(err)
	}
	newFile, err := utils.CompressImage(file)
	if err != nil {
		panic(err)
	}
	saveFile, _ := os.Create("./image/compress.jpeg")
	defer saveFile.Close()
	io.Copy(saveFile, newFile)
}
