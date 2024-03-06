package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	ApiModels "netdisk_in_go/APImodels"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
)

// GetUserStorage
// @Summary 获取用户存储容量
// @Produce json
// @Success 200 {object} string "存储容量"
// @Failure 400 {object} string "cookie校验失败"
// @Router /filetransfer/getstorage [get]
func GetUserStorage(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
	}
	// 获取用户信息
	ub, _ := models.FindUserByIdentity(utils.DB, uc.UserId)
	utils.RespOK(writer, 0, true, gin.H{
		"storageSize":      ub.StorageSize,
		"totalStorageSize": ub.TotalStorageSize,
	}, "存储容量")
}

// GetUserFileList 获取用户文件列表
func GetUserFileList(c *gin.Context) {
	writer := c.Writer
	// 校验cookie
	uc, isAuth := utils.CheckCookie(c)
	if !isAuth {
		utils.RespOK(writer, 999999, false, nil, "cookie校验失败")
		return
	}
	// 获取用户信息
	ub, isExist := models.FindUserByIdentity(utils.DB, uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
		return
	}

	// 获取请求参数
	var req ApiModels.UserFileListRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		utils.RespBadReq(writer, "请求参数不正确")
		return
	}
	var files []models.UserRepository
	var filesNum int
	if req.FileType == 0 {
		files, filesNum, err = models.FindFilesByPathAndPage(req.FilePath, ub.UserId, req.CurrentPage, req.PageCount)
	} else {
		files, filesNum, err = models.FindFilesByTypeAndPage(req.FileType, ub.UserId, req.CurrentPage, req.PageCount)
	}

	if err != gorm.ErrRecordNotFound && err != nil {
		utils.RespBadReq(writer, "参数错误")
		return
	}
	utils.RespOkWithDataList(writer, 0, files, int(filesNum), "文件列表")
}

// GetFileTree
func GetFileTree(c *gin.Context) {
	writer := c.Writer
	ub, err := models.GetUserFromCoookie(utils.DB, c)
	if err != nil {
		utils.RespOK(writer, 99999, false, nil, "用户校验失败")
	}
	// 存放文件树，返回给前端
	var root *models.FileTreeNode
	// 递归查询文件夹
	var ur []models.UserRepository
	res := utils.DB.Raw(`with RECURSIVE temp as
(
    select * from user_repository where file_name="/" AND user_id = ?
    union all
    select ur.* from user_repository as ur,temp t where ur.parent_id=t.user_file_id and ur.is_dir = 1 AND ur.deleted_at is NULL AND ur.user_id = ?
)
select * from temp;`, ub.UserId, ub.UserId).Find(&ur)
	fmt.Println(len(ur))
	if res.Error != nil {
		utils.RespOK(writer, ApiModels.DATABASEERROR, true, root, "查询文件树失败")
		return
	}
	if res.RowsAffected == 0 {
		utils.RespOK(writer, ApiModels.FILENOTEXIST, true, root, "查询文件树失败")
		return
	}
	root = models.BuildFileTreeFromDIr(ur)
	utils.RespOK(writer, 0, true, root, "成功")
}
