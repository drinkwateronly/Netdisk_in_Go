package sysconfig

import (
	"github.com/spf13/viper"
)

type ConfigModel struct {
	NetDiskConfig netDiskConfig `mapstructure:"netdisk"` // jwt密钥
	MysqlConfig   mysqlConfig   `mapstructure:"mysql"`   // mysql连接配置
	OfficeConfig  officeConfig  `mapstructure:"office"`  // onlyoffice配置
}

type netDiskConfig struct {
	Port        string    `mapstructure:"port"`        //
	JWTConfig   jwtConfig `mapstructure:"jwt"`         // jwt
	StorageSize uint64    `mapstructure:"storageSize"` //
}

type jwtConfig struct {
	Key            string `mapstructure:"key"`            //
	Issuer         string `mapstructure:"issuer"`         //
	CookieDuration int    `mapstructure:"cookieDuration"` //
}

// MysqlConfig 连接mysql的设置
type mysqlConfig struct {
	User   string `mapstructure:"user"`
	Pwd    string `mapstructure:"password"`
	DBName string `mapstructure:"dbname"`
	Host   string `mapstructure:"host"`
	Port   string `mapstructure:"port"`
}

// OnlyOffice 服务配置
type officeConfig struct {
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
