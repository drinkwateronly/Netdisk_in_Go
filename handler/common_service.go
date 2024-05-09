package handler

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/response"
)

// NoticeList
// @Summary 获取通知列表, 暂未使用
// @Tags unused
// @Accept json
// @Produce json
// @Router /notice/list [GET]
func NoticeList(c *gin.Context) {
	response.RespOK(c.Writer, 0, true, nil, "Notice")
	return
}

// GetCopyright
// @Summary 获取copyright, 暂未使用
// @Tags unused
// @Accept json
// @Produce json
// @Success 200 {object} api.CopyrightAPI{} ""
// @Router /param/grouplist [GET]
func GetCopyright(c *gin.Context) {
	response.RespOK(c.Writer, 0, true, api.CopyrightAPI{
		LicenseKey:        "",
		PrimaryDomainName: "",
		DomainChineseName: "",
		Project:           "",
		Company:           "",
		AuditDate:         "",
	}, "copyright")
	return
}
