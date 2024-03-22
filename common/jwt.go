package common

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"netdisk_in_go/sysconfig"
	"time"
)

type UserClaim struct {
	Username string
	Phone    string
	UserId   string
	jwt.RegisteredClaims
	// RegisteredClaims jwt.RegisteredClaims // 不要写成此种形式
}

// GenerateCookie 根据UserClaim所需字段签发cookie
func GenerateCookie(username, phone, userId string) (string, error) {
	// 从yaml文件获取相关配置
	issuer := sysconfig.Config.JWTConfig.Issuer
	key := sysconfig.Config.JWTConfig.Key
	cookieDuration := sysconfig.Config.JWTConfig.CookieDuration
	uc := UserClaim{
		Username: username,
		Phone:    phone,
		UserId:   userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			NotBefore: jwt.NewNumericDate(time.Now()),                                                  // 在该时间前生效
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(cookieDuration))), // 持续时长
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, uc)
	tokenString, err := token.SignedString([]byte(key)) // 根据jwt密钥签发
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ParseCookie 从token解析出UserClaim
func ParseCookie(token string) (*UserClaim, error) {
	// 新建userClaim结构体
	uc := new(UserClaim)
	// jwt.ParseWithClaims 输入 需要解析的JWT字符串、一个实现了jwt.Claims接口的结构体、用于提供验证签名所需的密钥的回调函数
	claims, err := jwt.ParseWithClaims(token, uc,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(sysconfig.Config.JWTConfig.Key), nil
		})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, errors.New("token is not invalid")
	}
	return uc, err
}
