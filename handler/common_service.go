package handler

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/api_models"
	"netdisk_in_go/utils"
)

// NoticeList
// @Summary 获取通知列表, 暂未使用
// @Accept json
// @Produce json
// @Router /notice/list [GET]
func NoticeList(c *gin.Context) {
	utils.RespOK(c.Writer, 0, true, nil, "Notice")
	return
}

// GetCopyright
// @Summary 获取copyright, 暂未使用
// @Accept json
// @Produce json
// @Success 200 {object} api_models.CopyrightAPI{} ""
// @Router /param/grouplist [GET]
func GetCopyright(c *gin.Context) {
	utils.RespOK(c.Writer, 0, true, api_models.CopyrightAPI{
		LicenseKey:        "",
		PrimaryDomainName: "",
		DomainChineseName: "",
		Project:           "",
		Company:           "",
		AuditDate:         "",
	}, "copyright")
	return
}
