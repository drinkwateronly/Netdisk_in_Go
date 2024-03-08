package utils

type ConfigModel struct {
	JWTkey           string `yaml:"JWTkey"`
	CookieExpireTime int    `yaml:"cookieExpiredTime"`
}
