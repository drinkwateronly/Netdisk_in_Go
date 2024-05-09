package main

import (
	"context"
	"encoding/json"
	"github.com/tencentyun/cos-go-sdk-v5"
	"log"
	"net/http"
	"net/url"
	config "netdisk_in_go/conifg"
	"netdisk_in_go/models"
	"netdisk_in_go/mq"
	"os"
)

// ProcessTransfer 文件上传OSS逻辑
func ProcessTransfer(msg []byte) bool {
	// 1.解析msg
	pubMsg := mq.TransferData{}
	err := json.Unmarshal(msg, &pubMsg)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	// 2.根据临时存储的文件路径，创建文件句柄
	_, err = os.Stat(pubMsg.TempLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	// 3.OSS上传
	u, _ := url.Parse("https://test-1306078367.cos.ap-guangzhou.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			// 通过环境变量获取密钥
			// 环境变量 SECRETID 表示用户的 SecretId，登录访问管理控制台查看密钥，https://console.cloud.tencent.com/cam/capi
			SecretID: os.Getenv("SecretID"), // 用户的 SecretId，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参见 https://cloud.tencent.com/document/product/598/37140
			// 环境变量 SECRETKEY 表示用户的 SecretKey，登录访问管理控制台查看密钥，https://console.cloud.tencent.com/cam/capi
			SecretKey: os.Getenv("SecretKey"), // 用户的 SecretKey，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参见 https://cloud.tencent.com/document/product/598/37140
		},
	})
	_, _, err = client.Object.Upload(
		context.Background(), pubMsg.FileHash, pubMsg.TempLocation, nil,
	)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	// 4.处理数据库
	res := models.DB.Model(&models.RepositoryPool{}).Where("file_id = ?", pubMsg.FildId).
		Update("oss", pubMsg.FileHash)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if res.RowsAffected == 0 {
		log.Println(err.Error())
		return false
	}

	log.Printf("consume: %s", pubMsg.FileHash)
	return true
}

func main() {
	// 开启一个消费者
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer,
	)
	log.Println("监听转移任务队列")
}
