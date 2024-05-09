package mq

// TransferData 转移队列中消息载体格式
type TransferData struct {
	FileHash      string
	FildId        string
	TempLocation  string
	DestLocation  string
	DestStoreType string
	DestType      int8 //存储类型，例如OSS存储、本地存储
}
