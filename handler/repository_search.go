package handler

import (
	"github.com/gin-gonic/gin"
	ApiModels "netdisk_in_go/api_models"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
)

// GetUserStorage
// @Summary 获取用户存储容量
// @Produce json
// @Success 200 {object} api_models.RespData{data=api_models.UserStorageReqAPI} ""
// @Failure 400 {object} string "cookie校验失败"
// @Router /filetransfer/getstorage [get]
func GetUserStorage(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}
	// 获取用户记录，此处默认用户一定存在，不校验isExist
	ub, _, err := models.FindUserByIdentity(utils.DB, uc.UserId)
	if err != nil {
		utils.RespOK(writer, 99999, true, nil, err.Error())
		return
	}
	utils.RespOK(writer, 0, true, ApiModels.UserStorageReqAPI{
		StorageSize:      ub.StorageSize,
		TotalStorageSize: ub.TotalStorageSize,
	}, "存储容量")
}

// GetUserFileList
// @Summary 根据文件类型或文件路径进行分页查询用户文件列表
// @Accept json
// @Produce json
// @Param req query api_models.UserFileListReqAPI true "请求"
// @Success 200 {object} api_models.RespDataList{dataList=[]api_models.UserFileListRespAPI} "文件列表"
// @Failure 400 {object} string "参数出错"
// @Router /file/getfilelist [GET]
func GetUserFileList(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}
	// 获取用户信息
	ub, isExist, err := models.FindUserByIdentity(utils.DB, uc.UserId)
	if !isExist {
		utils.RespOK(writer, ApiModels.USERNOTEXIST, false, nil, "用户不存在")
		return
	}
	// 绑定请求参数
	var req ApiModels.UserFileListReqAPI
	err = c.ShouldBindQuery(&req)
	if err != nil {
		utils.RespBadReq(writer, "请求参数不正确")
		return
	}
	// 查询文件记录
	var files []ApiModels.UserFileListRespAPI
	var filesTotalCount int
	if req.FileType == 0 {
		// 按文件路径分页查询
		files, filesTotalCount, err = models.FindFilesByPathAndPage(req.FilePath, ub.UserId, req.CurrentPage, req.PageCount)
	} else {
		// 按文件类型分页查询
		files, filesTotalCount, err = models.FindFilesByTypeAndPage(req.FileType, ub.UserId, req.CurrentPage, req.PageCount)
	}
	if err != nil {
		utils.RespOK(writer, ApiModels.DATABASEERROR, false, nil, err.Error())
		return
	}
	utils.RespOkWithDataList(writer, 0, files, filesTotalCount, "文件列表")
}

// GetFileTree
// @Summary 获取用户从根目录开始的文件树
// @Accept json
// @Produce json
// @Success 200 {object} api_models.RespData{data=api_models.UserFileTreeNode} "文件列表"
// @Failure 400 {object} string "参数出错"
// @Router /file/getfilelist [GET]
func GetFileTree(c *gin.Context) {
	writer := c.Writer
	ub, err := models.GetUserFromCoookie(utils.DB, c)
	if err != nil {
		utils.RespOK(writer, 99999, false, nil, "用户校验失败")
	}
	// 获取用户文件树
	var root *ApiModels.UserFileTreeNode
	root, err = models.BuildFileTree(ub.UserId)
	if err != nil {
		utils.RespOK(writer, ApiModels.DATABASEERROR, true, root, err.Error())
		return
	}
	utils.RespOK(writer, 0, true, root, "成功")
}
