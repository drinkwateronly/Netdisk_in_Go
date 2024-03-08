package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"netdisk_in_go/models"
	"netdisk_in_go/office_models"
	"netdisk_in_go/utils"
	"os"
)

// OnlyOffice
func PreviewOfficeFile(c *gin.Context) {
	writer := c.Writer
	// 从cookie获取用户信息
	cookie, _ := c.Cookie("token")
	uc, _ := utils.ParseCookie(cookie)

	// 获取用户信息
	ub, isExist, err := models.FindUserByIdentity(utils.DB, uc.UserId)
	if !isExist {
		utils.RespBadReq(writer, "用户不存在")
		return
	}
	// 处理请求参数
	json := make(map[string]interface{})
	err = c.ShouldBind(&json)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	userFileId := json["userFileId"].(string)
	// 查询用户文件基本信息
	rb, isExist := models.FindUserFileById(ub.UserId, userFileId)
	if !isExist {
		utils.RespBadReq(writer, "用户信息不存在")
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
	documentType, _ := utils.GetOfficeDocumentType(rb.ExtendName)

	utils.RespOK(writer, 200, true, office_models.NewData(user, cookie, document, documentType), "获取报告成功！")
	return
}

func OfficeFileDownload(c *gin.Context) {
	writer := c.Writer
	file, err := os.OpenFile("123.xls", os.O_RDONLY, 0777)
	defer file.Close()
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	utils.RespOK(writer, 0, true,
		struct {
			Error int `json:"error"`
		}{
			Error: 0,
		}, "下载成功")
}

func OfficeCallback(c *gin.Context) {
	// callback API: https://api.onlyoffice.com/editors/callback#changesurl
	writer := c.Writer

	// 获取post请求body中的参数
	var callbackHandler office_models.CallbackHandler
	c.ShouldBindJSON(&callbackHandler)
	fmt.Fprintf(gin.DefaultWriter, "%+v", callbackHandler) // 打印一下
	switch callbackHandler.Status {
	case 1: // document is being edited,
		// ignore this status
	case 2: // document is ready for saving,

	case 4: // document is closed with no changes
		// ignore this status
	case 6: // document is being edited, but the current document state is saved,
		fileTempUrl := callbackHandler.Url                    // presented when status = 2, 3, 6, 7
		fmt.Fprintf(gin.DefaultWriter, "%v", callbackHandler) // 打印一下
		// http 请求下载临时文件
		resp, err := http.Get(fileTempUrl)
		if err != nil {
			panic(err)
		}
		savePath := fmt.Sprintf("./repository/upload_file/%s", "test")

		saveFile, err := os.OpenFile(savePath, os.O_WRONLY, 0777)
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(saveFile, resp.Body)
		if err != nil {
			panic(err)
		}
		saveFile.Close()
		resp.Body.Close()
		// 处理 mysql
	default:
		/*
			default for error status:
				3: document saving error has occurred,
				7: error has occurred while force saving the document,
		*/
		ret, _ := json.Marshal(gin.H{
			"code":    0,
			"success": true,
			"error":   0,
		})
		_, err := writer.Write(ret)
		if err != nil {
			panic(err)
		}
		return
	}
	ret, _ := json.Marshal(gin.H{
		"code":    0,
		"success": true,
		"error":   0,
	})
	_, err := writer.Write(ret)
	if err != nil {
		panic(err)
	}

}

func OfficeFilePreview(c *gin.Context) {
	writer := c.Writer

	userFileId := c.Query("userFileId")
	cookie := c.Query("token")
	// 校验cookie
	uc, err := utils.ParseCookie(cookie)

	if err != nil {
		utils.RespOK(writer, 0, true, office_models.OfficeError{Error: 1}, "cookie校验失败")
		return
	}
	// 获取用户信息
	ub, isExist, err := models.FindUserByIdentity(utils.DB, uc.UserId)
	if !isExist {
		utils.RespOK(writer, 0, true, office_models.OfficeError{Error: 1}, "用户不存在")
		return
	}
	// 处理请求参数

	// 获取文件
	rp, isExist := models.FindRepFileByUserFileId(ub.UserId, userFileId)
	if !isExist {
		utils.RespOK(writer, 0, true, office_models.OfficeError{Error: 1}, "文件不存在")
		return
	}

	file, err := os.OpenFile(rp.Path, os.O_RDONLY, 0777)
	defer file.Close()
	_, err = io.Copy(c.Writer, file)
	if err != nil {
		utils.RespOK(writer, 0, true, office_models.OfficeError{Error: 1}, "出错")
		return
	}
	utils.RespOK(writer, 0, true, office_models.OfficeError{Error: 0}, "下载成功")

}
