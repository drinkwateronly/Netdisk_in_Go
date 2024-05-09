package common

import (
	"log"
	"os"
)

type netDiskLogger struct {
	*log.Logger
}

var NetDiskLogger netDiskLogger

func InitLogger() {
	file, _ := os.Create("logger.txt")
	NetDiskLogger = netDiskLogger{
		log.New(file, "example ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (NetDiskLogger *netDiskLogger) Info(s string) {
	NetDiskLogger.Println(s)
}
