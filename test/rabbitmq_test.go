package test

import (
	"github.com/streadway/amqp"
	config "netdisk_in_go/conifg"
	"testing"
)

func TestDialRabbitMq(t *testing.T) {
	amqpConn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, err = amqpConn.Channel()
	if err != nil {
		t.Fatal(err.Error())
	}
}
