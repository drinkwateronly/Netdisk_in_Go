package file_service

import (
	"github.com/gin-gonic/gin"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
)

// GetUserStorage
// @Summary 获取用户存储容量
// @Tags Files
// @Accept json
// @Produce json
// @Success 200 {object} response.RespData{data=api.UserStorageResp} "用户存储容量响应"
// @Router /filetransfer/getstorage [get]
func GetUserStorage(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	response.RespOKSuccess(writer, 0, api.UserStorageResp{
		StorageSize:      ub.StorageSize,
		TotalStorageSize: ub.TotalStorageSize,
	}, "存储容量")
}

// GetUserFileList
// @Summary 根据文件类型或文件路径进行分页查询用户文件列表
// @Accept json
// @Produce json
// @Param req query api.UserFileListReq true "请求"
// @Success 200 {object} response.RespDataList{dataList=[]api.UserFileListResp} "文件列表"
// @Router /file/getfilelist [GET]
func GetUserFileList(c *gin.Context) {
	writer := c.Writer
	ub := c.MustGet("userBasic").(*models.UserBasic)

	// 绑定请求参数
	var req api.UserFileListReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数不正确")
		return
	}

	// 查询文件记录
	var files []api.UserFileListResp
	// totalCount用于前端展示所有文件数量，而不是本次分页查询得到的文件数量
	var totalCount int
	if req.FileType == 0 {
		// 不选择文件类型时，按文件路径查找
		files, totalCount, err = models.PageQueryFilesByPath(req.FilePath, ub.UserId, req.CurrentPage, req.PageCount)
	} else {
		// 选择文件类型时，忽略文件路径
		files, totalCount, err = models.PageQueryFilesByType(req.FileType, ub.UserId, req.CurrentPage, req.PageCount)
	}
	if err != nil {
		response.RespOK(writer, response.DATABASEERROR, false, nil, err.Error())
		return
	}
	response.RespOkWithDataList(writer, response.Success, files, totalCount, "文件列表")
}

// GetFileTree
// @Summary 获取用户从根目录开始的文件树
// @Accept json
// @Produce json
// @Success 200 {object} response.RespData{data=api.UserFileTreeNode} "文件树根节点"
// @Router /file/getfiletree [GET]
func GetFileTree(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	// 获取用户文件树
	var root *api.UserFileTreeNode
	root, err := models.BuildFileTree(ub.UserId)
	if err != nil {
		response.RespOK(writer, response.DatabaseError, true, root, err.Error())
		return
	}
	response.RespOK(writer, 0, true, root, "成功")
}
