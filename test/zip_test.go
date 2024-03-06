package test

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestZipFile(t *testing.T) {

	zipFile, err := os.Create("./zip/test.zip")
	if err != nil {
		t.Fatal(err)
	}
	zipWriter := zip.NewWriter(zipFile)
	// 往zipWriter写入第一个文件
	w, err := zipWriter.Create("1.txt")
	if err != nil {
		t.Fatal(err)
		return
	}
	_, err = io.Copy(w, strings.NewReader("123"))
	if err != nil {
		t.Fatal(err)
		return
	}
	// 往zipWriter写入第二个文件
	w, err = zipWriter.Create("2.txt")
	if err != nil {
		t.Fatal(err)
		return
	}
	_, err = io.Copy(w, strings.NewReader("456"))
	if err != nil {
		t.Fatal(err)
		return
	}

	// 直接ReadFrom似乎不起作用
	zipFile2, err := os.Create("./zip/test2.zip")
	n, err := zipFile2.ReadFrom(zipFile)
	if err != nil {
		t.Fatal()
		return
	}
	//n, err := io.Copy(zipFile2, zipFile)
	//if err != nil {
	//	t.Fatal(err)
	//	return
	//}
	fmt.Println(n)
	zipFile2.Close()

	zipWriter.Close()
	zipFile.Close()

}

func TestFolderToZipFile(t *testing.T) {

	zipFile, err := os.Create("./zip/test.zip")
	if err != nil {
		t.Fatal(err)
	}
	zipWriter := zip.NewWriter(zipFile)
	// 往zipWriter写入第一个文件
	w, err := zipWriter.Create("chen/1.txt")
	if err != nil {
		t.Fatal(err)
		return
	}
	_, err = io.Copy(w, strings.NewReader("123"))
	if err != nil {
		t.Fatal(err)
		return
	}
	// 往zipWriter写入第二个文件
	w, err = zipWriter.Create("123/456/2.txt")
	if err != nil {
		t.Fatal(err)
		return
	}
	_, err = io.Copy(w, strings.NewReader("456"))
	if err != nil {
		t.Fatal(err)
		return
	}

	// 往zipWriter创建一个文件夹
	w, err = zipWriter.Create("789/456/")
	if err != nil {
		t.Fatal(err)
		return
	}

	zipWriter.Close()
	zipFile.Close()

}
