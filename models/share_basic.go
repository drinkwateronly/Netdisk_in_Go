package models

import (
	"gorm.io/gorm"
	"time"
)

type ShareBasic struct {
	gorm.Model
	UserId         string    `json:"userId"`
	Salt           string    `json:"salt"`
	ShareBatchId   string    `json:"shareBatchNum"`
	ShareType      uint8     `json:"shareType"`
	ExtractionCode string    `json:"extractionCode"`
	ExpireTime     time.Time `json:"expireTime"`
}

func (ShareBasic) TableName() string {
	return "share_basic"
}
