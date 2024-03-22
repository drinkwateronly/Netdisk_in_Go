package file_service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"netdisk_in_go/models"
	ApiModels "netdisk_in_go/models/api_models"
	"netdisk_in_go/utils"
	"os"
	"strings"
)

// FileDownload
// @Summary 文件单个下载接口
// @Produce json
// @Param userFileId query string true "单个用户文件标识符"
// @Param cookie query string true "Cookie"
// @Success 200 {object} string "服务器响应成功，根据响应code判断是否成功"
// @Failure 400 {object} string "参数出错"
// @Router /filetransfer/downloadfile [GET]
func FileDownload(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	userFileId := c.Query("userFileId")

	// 查询该文件的用户文件记录
	ur, isExist := models.FindUserFileById(models.DB, ub.UserId, userFileId)
	if !isExist {
		utils.RespOK(writer, ApiModels.FILENOTEXIST, false, nil, "文件记录不存在")
		return
	}

	// 情况1：下载文件不是文件夹，直接找到并发送文件。
	if ur.IsDir != 1 {
		// 获取中心存储文件记录
		rp, isExist := models.FindRepFileByUserFileId(ub.UserId, userFileId)
		if !isExist {
			utils.RespOK(writer, ApiModels.FILENOTEXIST, false, nil, "文件记录不存在")
			return
		}
		// 找到存储在服务器的文件
		fileInfo, err := os.Stat(rp.Path)
		if os.IsNotExist(err) {
			utils.RespOK(writer, ApiModels.FILENOTEXIST, false, nil, "存储文件丢失")
			return
		}
		// 传输文件
		file, err := os.OpenFile(rp.Path, os.O_RDONLY, 0777)
		defer file.Close()
		_, err = io.Copy(c.Writer, file)
		writer.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		if err != nil {
			utils.RespBadReq(writer, "出现错误")
			return
		}
		utils.RespOK(writer, 0, true, nil, "下载成功")
		return
	}

	// 情况2：下载文件是文件夹，需要找到文件，并按照文件在文件夹的存放位置进行压缩。
	// 根据用户文件记录生成zip文件，并返回该文件存储在服务器的路径
	zipFilePath, err := models.GenZipFromUserRepos(*ur)
	if err != nil {
		utils.RespOK(writer, 99999, false, nil, "创建zip文件失败")
		return
	}
	// 检查文件是否存在
	_, err = os.Stat(zipFilePath)
	if os.IsNotExist(err) {
		// zip文件不存在，返回错误信息
		utils.RespOK(writer, 99999, false, nil, "zip文件不存在")
		return
	}
	// zip文件存在，打开文件
	zipFile, err := os.OpenFile(zipFilePath, os.O_RDONLY, 0777)
	defer zipFile.Close() // defer 关闭文件
	if err != nil {
		utils.RespOK(writer, 99999, false, nil, "无法打开zip文件")
		return
	}
	// 发送文件
	_, err = io.Copy(writer, zipFile)
	if err != nil {
		utils.RespOK(writer, 99999, false, nil, "文件io出错")
		return
	}
	utils.RespOK(writer, 0, true, nil, "下载成功")
	return
}

// FileDownloadInBatch
// @Summary 文件批量下载接口
// @Produce json
// @Param userFileIds query string true "多个用户文件标识符，以逗号隔开"
// @Param cookie query string true "Cookie"
// @Success 200 {object} string "服务器响应成功，根据响应code判断是否成功"
// @Failure 400 {object} string "参数出错"
// @Router /filetransfer/batchDownloadFile [GET]
func FileDownloadInBatch(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	// 获取查询参数，并分割出文件id切片
	userFileIds := strings.Split(c.Query("userFileIds"), ",")
	if len(userFileIds) == 0 {
		utils.RespBadReq(writer, "参数不正确")
		return
	}
	// 找到根据文件id找到用户文件记录
	userRepos, _ := models.FindUserFileByIds(ub.UserId, userFileIds)
	// 根据用户文件记录生成zip压缩文件（核心功能）
	zipFilePath, err := models.GenZipFromUserRepos(*userRepos...)
	if err != nil {
		return
	}
	// 查询zip文件信息
	zipFileInfo, err := os.Stat(zipFilePath)
	// zip文件不存在，返回错误信息
	if os.IsNotExist(err) {
		utils.RespBadReq(writer, "出现错误")
		return
	}
	// zip文件存在，打开文件
	zipFile, err := os.Open(zipFilePath)
	_, err = io.Copy(c.Writer, zipFile)
	if err != nil {
		utils.RespBadReq(writer, "出现错误")
		return
	}

	writer.Header().Set("Content-Length", fmt.Sprintf("%d", zipFileInfo.Size()))
	writer.Header().Set("Content-Type", fmt.Sprintf("application/x-zip-compressed"))
	utils.RespOK(writer, 0, true, nil, "下载成功")
	return
}
