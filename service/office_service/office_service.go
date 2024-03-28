package office_service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"netdisk_in_go/office_models"
	"os"
)

// PreviewOfficeFile
// 预览office文件
func PreviewOfficeFile(c *gin.Context) {
	writer := c.Writer
	// 从cookie获取用户信息
	cookie, _ := c.Cookie("token")
	uc, _ := common.ParseCookie(cookie)

	// 获取用户信息
	ub, isExist, err := models.FindUserByIdentity(models.DB, uc.UserId)
	if !isExist {
		response.RespBadReq(writer, "用户不存在")
		return
	}
	// 处理请求参数
	json := make(map[string]interface{})
	err = c.ShouldBind(&json)
	if err != nil {
		response.RespBadReq(writer, "出现错误")
		return
	}
	userFileId := json["userFileId"].(string)
	// 查询用户文件基本信息
	rb, err := models.FindUserFileById(models.DB, ub.UserId, userFileId)
	if err != nil {
		response.RespBadReq(writer, "用户信息不存在")
		return
	}

	// 根据API返回必要的信息
	// https://api.onlyoffice.com/editors/config/editor
	document := office_models.Document{
		FileType: rb.ExtendName, // 文件拓展名
		Info: office_models.Info{
			Owner:  "Me",
			Upload: rb.UpdatedAt.Format("Mon Jan 02 2006"),
		},
		Key:         "",                                                                 // todo: 看看有无其他用法
		Permissions: office_models.DefaultPermissions,                                   // 使用默认的 Permissions
		Title:       rb.FileName + "." + rb.ExtendName,                                  // 文件完整名
		Url:         fmt.Sprintf(office_models.PreviewUrlFormat, rb.UserFileId, cookie), // 文件预览链接
		UserFileId:  rb.FileId,                                                          // 文件id
	}
	user := office_models.User{
		Id:    uc.UserId,
		Name:  uc.UserId,
		Group: "",
	}
	documentType, _ := filehandler.GetOfficeDocumentType(rb.ExtendName)

	response.RespOK(writer, 200, true, office_models.NewData(user, cookie, document, documentType), "获取报告成功！")
	return
}

func OfficeFileDownload(c *gin.Context) {
	writer := c.Writer
	file, err := os.OpenFile("123.xls", os.O_RDONLY, 0777)
	defer file.Close()
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		response.RespBadReq(writer, "出现错误")
		return
	}
	response.RespOK(writer, 0, true,
		struct {
			Error int `json:"error"`
		}{
			Error: 0,
		}, "下载成功")
}

// OfficeCallback
// @Summary OnlyOffice文件编辑的回调接口
// @Description
// @Accept json
// @Produce json
// @Param req query api.CallbackHandler true "请求"
// @Success 200 {object} response.RespData{data=api.OfficeErrorResp} "响应，成功时为文件"
// @Router /office/preview [GET]
func OfficeCallback(c *gin.Context) {
	// callback API: https://api.onlyoffice.com/editors/callback#changesurl
	writer := c.Writer

	// 获取post请求body中的参数
	var callbackHandler api.OfficeCallbackReq
	err := c.ShouldBindJSON(&callbackHandler)
	fmt.Printf("%+v\n", callbackHandler)
	fmt.Printf("%d\n", callbackHandler.Status)
	//if err != nil {
	//	return
	//}
	//fmt.Fprintf(gin.DefaultWriter, "%+v", callbackHandler) // 打印一下
	//
	//switch callbackHandler.Status {
	//case 1: // document is being edited,
	//	// ignore this status
	//case 2: // document is ready for saving,
	//
	//case 4: // document is closed with no changes
	//	// ignore this status
	//case 6: // document is being edited, but the current document state is saved,
	//	fileTempUrl := callbackHandler.Url                    // presented when status = 2, 3, 6, 7
	//	fmt.Fprintf(gin.DefaultWriter, "%v", callbackHandler) // 打印一下
	//	// http 请求下载临时文件
	//	resp, err := http.Get(fileTempUrl)
	//	if err != nil {
	//		panic(err)
	//	}
	//	savePath := fmt.Sprintf("./repository/upload_file/%s", "test")
	//
	//	saveFile, err := os.OpenFile(savePath, os.O_WRONLY, 0777)
	//	if err != nil {
	//		panic(err)
	//	}
	//	_, err = io.Copy(saveFile, resp.Body)
	//	if err != nil {
	//		panic(err)
	//	}
	//
	//	// 处理 mysql
	//default:
	//	/*
	//		default for error status:
	//			3: document saving error has occurred,
	//			7: error has occurred while force saving the document,
	//	*/
	//	//ret, _ := json.Marshal(gin.H{
	//	//	"error": 0,
	//	//})
	//	//_, err := writer.Write(ret)
	//	//if err != nil {
	//	//	panic(err)
	//	//}
	//	//return
	//}
	//ret, _ := json.Marshal(api.OfficeErrorResp{Error: 0})
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write([]byte("{\"error\":0}"))
	if err != nil {
		panic(err)
	}

}

// OfficeFilePreview
// @Summary onlyoffice文件预览
// @Description
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
	_, err = writer.Write([]byte("{\"error\":0}"))
}
