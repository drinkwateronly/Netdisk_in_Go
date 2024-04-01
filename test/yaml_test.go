package test

import (
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io"
	"netdisk_in_go/sysconfig"
	"os"
	"testing"
)

func TestYamlParse(t *testing.T) {
	f, err := os.Open("../config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var conf sysconfig.ConfigModel

	var data []byte
	data, err = io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	err = yaml.Unmarshal(data, &conf)

	if err != nil {
		t.Fatal(err)
	}

}

func TestViper(t *testing.T) {
	viper.New()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../")
	var conf sysconfig.ConfigModel

	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	} // 加载配置文件出错

	if err := viper.Unmarshal(&conf); err != nil {
		t.Fatal(err)
	}
}
