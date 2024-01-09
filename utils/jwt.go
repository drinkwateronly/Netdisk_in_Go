package utils

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type UserClaim struct {
	Phone                string
	UserId               string
	jwt.RegisteredClaims // 不要写成RegisteredClaims jwt.RegisteredClaims
}

func GenerateToken(phone string, userId string, expireTime int) (string, error) {
	uc := UserClaim{
		Phone:  phone,
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "CHENJIE",
			NotBefore: jwt.NewNumericDate(time.Now()), // 在该时间前生效
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(expireTime))),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, uc)
	tokenString, err := token.SignedString([]byte("jwt-key")) // todo:key放到配置文件中
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
			return []byte("jwt-key"), nil
		})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, errors.New("token is not invalid")
	}
	return uc, err
}

// ParseCookieFromRequest 从Context中获取cookie并解析
func ParseCookieFromRequest(c *gin.Context) (*UserClaim, error) {
	// 判断请求有无带token
	token, err := c.Cookie("token")
	if err != nil {
		return nil, errors.New("no token in request")
	}
	return ParseCookie(token)
}

// CheckCookie 解析cookie，并判断是否有效
func CheckCookie(c *gin.Context) (*UserClaim, bool) {
	uc, err := ParseCookieFromRequest(c)
	if err != nil {
		return nil, false
	}
	return uc, true
}
