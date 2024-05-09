package mq

import (
	"github.com/streadway/amqp"
	"log"
	config "netdisk_in_go/conifg"
)

var amqpConn *amqp.Connection
var amqpChan *amqp.Channel

func init() {
	initChannel()
}

func initChannel() bool {
	// 1.判断channel是否已经创建
	if amqpChan != nil {
		return true
	}
	// 2.获取rabbitmq的连接
	var err error
	amqpConn, err = amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Println("err.Error()" + err.Error())
		return false
	}
	// 3.打开一个channel，用于消息的发布和接收
	amqpChan, err = amqpConn.Channel()
	if err != nil {
		log.Println("amqpConn.Channel():" + err.Error())
		return false
	}
	return true
}

// Publish 发布消息到消息队列
func Publish(exchangeName, routingKey string, msg []byte) bool {
	// 1.初始化channel，若channel已初始化，则直接使用
	if !initChannel() {
		log.Println("InitChannel")
		return false
	}
	// 2.执行消息发布动作
	err := amqpChan.Publish(exchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: "text/plain", // 明文
		Body:        msg,
	})
	if err != nil {
		log.Println("amqpChan.Publish" + err.Error())
		return false
	}
	return true
}
