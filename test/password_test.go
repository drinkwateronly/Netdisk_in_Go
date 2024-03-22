package test

import (
	"fmt"
	"netdisk_in_go/common"
	"testing"
)

func TestPassword(t *testing.T) {
	rawPassword := "19990414"
	salt := "1298498081"
	fmt.Println(salt)
	password := common.MakePassword(rawPassword, salt)
	fmt.Println(password)
	if !common.ValidatePassword(rawPassword, salt, password) {
		t.Error("password validation failed")
	}
}

func TestCode(t *testing.T) {
	for i := 0; i < 5; i++ {
		fmt.Println(common.GenerateRandCode())
	}
}
