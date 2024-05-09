package config

const (
	StoreOSS   = 0
	StoreLocal = iota
)

const (
	// AsyncTransferEnable 是否开启文件异步转移
	AsyncTransferEnable = true
	// RabbitURL rabbitmq服务的入口地址与端口
	RabbitURL = "amqp://guest:guest@172.31.226.34:5672/"
	// TransExchangeName 用于文件上传OSS的交换机
	TransExchangeName = "uploadserver.trans"
	// TransOSSQueueName OSS转移队列名
	TransOSSQueueName = "uploadserver.trans.oss"
	// TransOSSErrQueueName OSS上传失败后写入另一个队列
	TransOSSErrQueueName = "uploadserver.trans.oss.err"
	// TransOSSRoutingKey key
	TransOSSRoutingKey = "oss"
)
