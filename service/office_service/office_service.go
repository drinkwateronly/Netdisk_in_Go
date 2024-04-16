package office_service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"io"
	"net/http"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"netdisk_in_go/common/office_models"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"os"
)

// PrepareOnlyOffice
// @Summary office文件预览与编辑前的准备接口
// @Description 点击office文件时，该接口用于获取文件信息、文件预览接口、后端回调接口以及一些OnlyOffice的基本设置，为后续编辑文件做准备
// @Tags office
// @Accept json
// @Produce json
// @Param req body api.PrepareOnlyOfficeReq true "请求"
// @Success 200 {object} response.RespData{data=office_models.OnlyOfficeConfig} "响应"
// @Router /office/previewofficefile [POST]
func PrepareOnlyOffice(c *gin.Context) {
	writer := c.Writer
	// 从cookie获取用户信息
	cookie, err := c.Cookie("token")
	if err != nil {
		response.RespUnAuthorized(writer)
		return
	}
	uc, err := common.ParseCookie(cookie)
	if err != nil {
		response.RespUnAuthorized(writer)
		return
	}
	// 获取用户信息
	ub, isExist, err := models.FindUserByIdentity(models.DB, uc.UserId)
	if !isExist {
		response.RespBadReq(writer, "用户不存在")
		return
	}
	// 解析请求
	var req api.PrepareOnlyOfficeReq
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.RespBadReq(writer, "出现错误")
		return
	}
	// 查询用户文件
	ur, err := models.FindUserFileById(models.DB, ub.UserId, req.UserFileId)
	if err != nil {
		response.RespOKFail(writer, response.FileNotExist, "文件不存在")
		return
	}
	// 根据API返回必要的信息, 参考：https://api.onlyoffice.com/editors/config/editor
	document := office_models.Document{
		FileType: ur.ExtendName, // 文件拓展名
		Info: office_models.Info{
			Owner:  "Me",
			Upload: ur.UpdatedAt.Format("Mon Jan 02 2006"),
		},
		Key:         common.GenerateUUID(),                                              // https://forum.onlyoffice.com/t/how-to-manage-document-key-correctly/1536
		Permissions: office_models.DefaultPermissions,                                   // 使用默认的 Permissions
		Title:       ur.FileName + "." + ur.ExtendName,                                  // 文件完整名
		Url:         fmt.Sprintf(office_models.PreviewUrlFormat, ur.UserFileId, cookie), // 文件预览链接
		UserFileId:  ur.FileId,                                                          // 文件id
	}
	user := office_models.User{
		Id:   uc.UserId,
		Name: uc.Username,
	}
	documentType, ok := filehandler.GetOfficeDocumentType(ur.ExtendName)
	if !ok {
		// 未找到文件类型
		response.RespOKFail(writer, response.NotSupport, "文件类型不支持OnlyOffice服务")
		return
	}
	response.RespOKSuccess(writer, response.OfficePrepareSuccess, office_models.NewOnlyOfficeConfig(user, cookie, document, documentType), "获取报告成功！")
	return
}

// OfficeCallback
// @Summary OnlyOffice回调接口
// @Description 对onlyoffice服务中所编辑的文件进行保存
// @Tags office
// @Accept json
// @Produce json
// @Param req body api.OfficeCallbackReq true "请求"
// @Success 200 {object} response.RespData{data=api.OfficeErrorResp} "响应，成功时为文件"
// @Router /office/callback [POST]
func OfficeCallback(c *gin.Context) {
	// callback API: https://api.onlyoffice.com/editors/callback#changesurl
	writer := c.Writer

	// 获取post请求body中的参数
	var callbackHandler api.OfficeCallbackReq
	err := c.ShouldBindJSON(&callbackHandler)
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
		return
	}

	switch callbackHandler.Status {
	case 1, 4:
		/* ignore status
		1: document is being edited,
		4: document is closed with no changes,
		*/
	case 2, 6:
		/* saving status
		2: document is ready for saving,
		6: document is being edited, but the current document state is saved,
		*/
		// http 请求onlyoffice端下载修改后临时文件
		fileTempUrl := callbackHandler.Url
		fmt.Fprintf(gin.DefaultWriter, "%v", callbackHandler) // 打印一下
		resp, err := http.Get(fileTempUrl)
		if err != nil {
			response.RespOKFail(writer, response.FileIOError, "临时文件下载错误")
			return
		}
		// 文件修改，生成新的中心存储池id
		fileId := common.GenerateUUID()
		savePath := fmt.Sprintf("./repository/upload_file/%s", fileId)
		saveFile, err := os.OpenFile(savePath, os.O_WRONLY, 0777)
		if err != nil {
			response.RespOKFail(writer, response.FileIOError, "临时文件下载错误")
			return
		}
		n, err := io.Copy(saveFile, resp.Body)
		if err != nil {
			response.RespOKFail(writer, response.FileIOError, "临时文件下载错误")
			return
		}

		// 文件下载完毕 处理mysql
		models.DB.Transaction(func(tx *gorm.DB) error {
			rp := models.RepositoryPool{
				FileId: fileId,
				Hash:   "",
				Size:   uint64(n),
				Path:   savePath,
			}
			err = tx.Create(&rp).Error
			if err != nil {
				response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
				return err
			}
			err = tx.Model(&models.UserRepository{}).Where("user_file_id = ? AND user_id = ?", callbackHandler.Key, callbackHandler.Users[0]).
				Updates(&models.UserRepository{FileId: fileId}).Error // 只更新非0列
			if err != nil {
				response.RespOKFail(writer, response.DatabaseError, "DatabaseError")
				return err
			}
			return nil
		})
	default:
		/* default for error status:
		3: document saving error has occurred,
		7: error has occurred while force saving the document,
		*/
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("{\"error\":0}"))
}

// OfficeFilePreview
// @Summary onlyoffice文件预览
// @Description
// @Tags office
// @Accept json
// @Produce json
// @Param req query api.OfficeFilePreviewReq true "请求"
// @Success 200 {object} response.RespData{data=api.OfficeErrorResp} "响应，成功时为文件"
// @Router /office/preview [GET]
func OfficeFilePreview(c *gin.Context) {
	writer := c.Writer
	// 处理请求参数
	req := api.OfficeFilePreviewReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespOK(writer, response.Unauthorized, false, api.OfficeErrorResp{Error: 1}, "参数错误")
		return
	}
	// 校验cookie
	uc, err := common.ParseCookie(req.Cookie)
	if err != nil {
		response.RespOK(writer, response.Unauthorized, false, api.OfficeErrorResp{Error: 1}, "cookie校验失败")
		return
	}
	// 获取用户信息
	ub, isExist, err := models.FindUserByIdentity(models.DB, uc.UserId)
	if !isExist {
		response.RespOK(writer, response.UserNotExist, true, api.OfficeErrorResp{Error: 1}, "用户不存在")
		return
	}
	// 获取文件
	rp, isExist := models.FindRepFileByUserFileId(models.DB, ub.UserId, req.UserFileId)
	if !isExist {
		response.RespOK(writer, response.FileNotExist, true, api.OfficeErrorResp{Error: 1}, "文件不存在")
		return
	}
	// 打开文件并发送
	file, err := os.OpenFile(rp.Path, os.O_RDONLY, 0777)
	defer file.Close()
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		response.RespOK(writer, response.FileIOError, true, api.OfficeErrorResp{Error: 1}, "出错")
		return
	}
}
