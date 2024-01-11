package test

import (
	"fmt"
	"log"
	"os/exec"
	"testing"
)

func TestExec(t *testing.T) {
	cmd := exec.Command("ls", "-l") // 这里以 "ls -l" 为例子，可根据需要修改成任意命令及参数

	output, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	// 将字节切片转换为字符串
	result := string(output[:])
	fmt.Println(result)
}
