package sysconfig

import (
	"github.com/spf13/viper"
)

type ConfigModel struct {
	CookieExpireTime int          `mapstructure:"cookieExpiredTime"` // cookie时长
	JWTConfig        JWTConfig    `mapstructure:"JWT"`               // jwt密钥
	MysqlConfig      MysqlConfig  `mapstructure:"mysql"`             // mysql连接配置
	OfficeConfig     OfficeConfig `mapstructure:"office"`            // onlyoffice配置
}

type JWTConfig struct {
	Key            string `mapstructure:"key"`            //
	Issuer         string `mapstructure:"issuer"`         //
	CookieDuration int    `mapstructure:"cookieDuration"` //
}

// MysqlConfig 连接mysql的设置
type MysqlConfig struct {
	User   string `mapstructure:"user"`
	Pwd    string `mapstructure:"password"`
	DBName string `mapstructure:"dbname"`
	Host   string `mapstructure:"host"`
	Port   string `mapstructure:"port"`
}

type OfficeConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

var Config ConfigModel

func InitConfig() {
	viper.New()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")

	err := viper.ReadInConfig() // 加载配置文件出错
	if err != nil {
		panic(err)
	}
	if err := viper.Unmarshal(&Config); err != nil {
		panic(err)
	}

}
