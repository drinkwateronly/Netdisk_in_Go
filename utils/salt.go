package utils

// MakeSalt 随机生成盐，todo:换成更好的算法
func MakeSalt() string {
	return GenerateUUID()
}

// MakePassword 生成记录在数据库的密码
func MakePassword(rawPassword, salt string) string {
	return Md5Encode(rawPassword + salt)
}

// ValidatePassword 校验密码
func ValidatePassword(rawPassword, salt, Md5Password string) bool {
	return Md5Encode(rawPassword+salt) == Md5Password
}
