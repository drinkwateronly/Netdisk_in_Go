package utils

import (
	"math/rand"
	"time"
)

// GenerateRandCode 返回一个长度为6的纯数字随机码用于网盘生成分享密码
func GenerateRandCode() string {
	codeLen := 6 // 长度为6
	// 只用时间戳时，相同时间戳下会生成一样的code
	rand.Seed(time.Now().UnixNano() + int64(rand.Intn(999999)))
	nums := [10]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	code := make([]byte, codeLen)
	for i := range code {
		code[i] = nums[rand.Intn(10)]
	}
	//return fmt.Sprintf("%6d", rand.Intn(999999))
	return string(code)
}
