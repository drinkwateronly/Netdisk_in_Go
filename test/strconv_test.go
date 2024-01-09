package test

import (
	"fmt"
	"strconv"
	"testing"
)

func TestAtoI(t *testing.T) {
	num, err := strconv.Atoi("")
	fmt.Println(num)
	if err != nil {
		t.Fatal(err)
	}
}
