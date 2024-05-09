package mq

import (
	"log"
)

var done chan bool

func StartConsume(qName, cName string, callback func(msg []byte) bool) {
	// 1.获得无缓冲channel类型的消息信道
	msgs, err := amqpChan.Consume(
		qName, // 队列名称
		cName, // 消费者名称
		true,  // 自动回复生产者
		false, // 可能有多个消费者，竞争机制派发消息
		false, //
		false, //
		nil,   //
	)
	if err != nil {
		log.Println("amqpChan.Consume:" + err.Error())
		return
	}
	// 2.循环从channel获取队列的消息
	done = make(chan bool)
	go func() {
		// 3.获取消息后调用callback方法处理消息
		for msg := range msgs {
			if !callback(msg.Body) {
				// todo：callback处理失败，写到另一个队列等待重试
			}
		}
	}()
	// 阻塞
	<-done
	// 关闭amqpChan
	defer amqpChan.Close()
	return
}
