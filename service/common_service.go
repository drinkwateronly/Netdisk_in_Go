package service

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/utils"
)

func NoticeList(c *gin.Context) {
	utils.RespOK(c.Writer, 0, true, nil, "Notice")
}

func GetCopyright(c *gin.Context) {
	utils.RespOK(c.Writer, 0, true, gin.H{
		"licenseKey":        nil,
		"primaryDomainName": nil,
		"domainChineseName": "深圳大学无线网络研究小组",
		"project":           "网盘",
		"company":           "深圳大学无线网络研究小组",
		"auditDate":         "2023-12-01",
	}, "copyright")
}

///param/grouplist?groupName=copyright
