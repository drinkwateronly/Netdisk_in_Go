package test

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"testing"
)

type Config struct {
	JWTkey           string `yaml:"JWTkey"`
	CookieExpireTime int    `yaml:"cookieExpiredTime"`
}

func TestYamlParseTest(t *testing.T) {
	f, err := os.Open("../config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var config Config

	var data []byte
	data, err = io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	err = yaml.Unmarshal(data, &config)

	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(config.JWTkey)
	fmt.Println(config.CookieExpireTime)
}
