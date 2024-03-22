package recovery_service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"strings"
)

//  回收站文件相关接口

// GetRecoveryFileList
// @Summary 获取回收站文件列表
// @Accept json
// @Produce json
// @Param cookie query string true "Cookie" // 并非query参数
// @Success 200 {object} api_models.RespDataList{datalist=[]api_models.RecoveryListRespAPI} "服务器响应成功，根据响应code判断是否成功"
// @Failure 400 {object} string "参数出错"
// @Router /recoveryfile/list [GET]
func GetRecoveryFileList(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	var recoveryFiles []api.RecoveryListRespAPI
	res := models.DB.Model(models.RecoveryBasic{}).Where("user_id = ?", ub.UserId).Scan(&recoveryFiles)
	if res.Error != nil {
		response.RespOK(writer, 99999, false, nil, "")
		return
	}
	response.RespOkWithDataList(writer, 0, recoveryFiles, len(recoveryFiles), "文件列表")
}

// DelRecoveryFile
// @Summary 删除单个回收站文件
// @Accept json
// @Produce json
// @Param userFileId body string true "用户文件id"
// @Param cookie query string true "Cookie" // 并非query参数
// @Success 200 {object} api_models.RespData{} ""
// @Failure 400 {object} string "参数出错"
// @Router /recoveryfile/deleterecoveryfile [POST]
func DelRecoveryFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)

	// 绑定post载荷的json格式参数
	var r api.DelRecoveryReqAPI
	err := c.ShouldBindJSON(&r)
	if err != nil {
		response.RespBadReq(writer, "参数错误")
		return
	}
	if err := models.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, r.UserFileId).
		Delete(&models.RecoveryBasic{}).Error; err != nil {
		response.RespOK(writer, 0, false, nil, "文件删除失败")
		return
	}
	response.RespOkWithDataList(writer, 0, nil, 0, "删除成功")
}

// DelRecoveryFilesInBatch
// @Summary 删除多个回收站文件
// @Accept json
// @Produce json
// @Param userFileId body string true "用户文件id"
// @Param cookie query string true "Cookie" // 并非query参数
// @Success 200 {object} api_models.RespData{} ""
// @Failure 400 {object} string "参数出错"
// @Router /recoveryfile/deleterecoveryfile [POST]
func DelRecoveryFilesInBatch(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)

	var r api.DelRecoveryFilesInBatchReq
	err := c.ShouldBindJSON(&r)
	if err != nil {
		response.RespBadReq(writer, "参数错误")
		return
	}
	userFileIdList := strings.Split(r.UserFileIds, ",")
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		for _, userFileId := range userFileIdList {
			if err := models.DB.Where("user_id = ? AND user_file_id = ?", ub.UserId, userFileId).
				Delete(&models.RecoveryBasic{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.RespOK(writer, 0, false, nil, "清空回收站失败")
		return
	}
	response.RespOkWithDataList(writer, 0, nil, 0, "清空回收站成功")
}
