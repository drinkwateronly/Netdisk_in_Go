package test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestAtoI(t *testing.T) {
	num, err := strconv.Atoi("")
	fmt.Println(num)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStringSplit(t *testing.T) {
	strList := []string{"123\\456\\abc.jpg", "abc.jpg"}
	for _, str := range strList {
		fileName := "abc.jpg"
		splitStr := strings.Split(str[:len(str)-len(fileName)-1], "\\")
		fmt.Println(splitStr, len(splitStr))
	}

}
