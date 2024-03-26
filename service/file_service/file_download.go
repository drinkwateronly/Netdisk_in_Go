package file_service

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"os"
	"strings"
	"time"
)

// FileDownload
// @Summary 单个文件下载接口
// @Tags filetransfer
// @Accept json
// @Produce json
// @Param req query api.FileDownloadReq true "请求"
// @Router /filetransfer/downloadfile [GET]
func FileDownload(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)

	// 参数绑定
	req := api.FileDownloadReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数不正确")
		return
	}
	userFileId := req.UserFileId

	// 查询该文件的用户文件记录
	ur, err := models.FindUserFileById(models.DB, ub.UserId, userFileId)
	if err != nil {
		response.RespOK(writer, response.FileNotExist, false, nil, "文件记录不存在")
		return
	}
	// 文件存放路径
	var savePath string
	if ur.IsDir != 1 { // 情况1：下载文件不是文件夹，直接发送此文件。
		// 获取中心存储池文件记录
		rp, isExist := models.FindRepFileByUserFileId(models.DB, ub.UserId, userFileId)
		if !isExist {
			response.RespOKFail(writer, response.FileNotExist, "文件记录不存在")
			return
		}
		savePath = rp.Path
	} else { // 情况2：下载文件是文件夹，需要找到文件夹内部所有文件，并按照相对位置存放并压缩。
		// 根据请求的文件夹记录，生成zip文件，并返回该文件存储在服务器的路径
		savePath, err = models.GenZipFromUserRepos(ur)
		if err != nil {
			response.RespOK(writer, response.GenZipError, false, nil, "创建zip文件失败")
			return
		}
	}

	/*
		// 文件是否存在
		fileInfo, err := os.Stat(savePath)
		if os.IsNotExist(err) {
			// zip文件不存在，返回错误信息
			response.RespOKFail(writer, response.SaveFileNotExist, "文件不存在")
			return
		}
	*/
	// 打开存储在服务器的文件
	file, err := os.OpenFile(savePath, os.O_RDONLY, 0777)
	defer file.Close() // defer 关闭文件
	if err != nil {
		response.RespOKFail(writer, response.SaveFileNotExist, "文件不存在")
		return
	}

	// 在文件传输前设置Content-Length
	// todo：偶尔出现http: wrote more than the declared Content-Length错误，暂时注释
	// writer.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// 文件流式传输，不会将整个文件放到内存
	_, err = io.Copy(writer, file)
	if err != nil {
		response.RespOKFail(writer, response.FileIOError, "文件io出错")
		return
	}
	//response.RespOKSuccess(writer, response.Success, nil, "下载成功")
	return
}

// FileDownloadInBatch
// @Summary 文件批量下载接口
// @Tags filetransfer
// @Accept json
// @Produce json
// @Param req query api.FileDownloadInBatchReq true "请求"
// @Router /filetransfer/batchDownloadFile [GET]
func FileDownloadInBatch(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)

	// 获取查询参数，并分割出文件id切片
	req := api.FileDownloadInBatchReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数不正确")
		return
	}
	// 以逗号分割
	userFileIds := strings.Split(req.UserFileIds, ",")
	if len(userFileIds) == 0 {
		response.RespBadReq(writer, "请求参数不正确")
		return
	}

	// 找到根据文件id找到用户文件记录
	userRepos, _ := models.FindUserFilesByIds(models.DB, ub.UserId, userFileIds)

	// 根据用户文件记录生成zip压缩文件（核心功能）
	zipFilePath, err := models.GenZipFromUserRepos(userRepos...)
	if err != nil {
		return
	}
	/*
		// zip文件信息
		_, err = os.Stat(zipFilePath)
		// zip文件不存在
		if os.IsNotExist(err) {
			response.RespOKFail(writer, response.FileNotExist, "zip文件不存在")
			return
		}
	*/
	// 打开文件
	zipFile, err := os.Open(zipFilePath)
	if os.IsNotExist(err) {
		response.RespOKFail(writer, response.FileNotExist, "zip文件不存在")
		return
	}
	// 设置header
	// todo：偶尔出现http: wrote more than the declared Content-Length错误，暂时注释
	// writer.Header().Set("Content-Length", fmt.Sprintf("%d", zipFileInfo.Size()))
	writer.Header().Set("Content-Type", "application/x-zip-compressed")
	// 流式传输
	_, err = io.Copy(c.Writer, zipFile)
	if err != nil {
		response.RespOKFail(writer, response.FileIOError, "zip文件不存在")
		return
	}
	//response.RespOK(writer, response.Success, true, nil, "下载成功")
	return
}

// FilePreview
// @Summary 文件预览
// @Tags filetransfer
// @Produce json
// @Accept json
// @Param req query api.FileDownloadInBatchReq true "请求"
// @Router /filetransfer/preview [GET]
func FilePreview(c *gin.Context) {
	writer := c.Writer

	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)

	// 获取用户信息
	ub, isExist, err := models.FindUserByIdentity(models.DB, ub.UserId)
	if !isExist {
		response.RespBadReq(writer, "用户不存在")
		return
	}

	// 处理请求参数
	req := api.FilePreviewReq{}
	err = c.ShouldBindQuery(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
	}
	userFileId := req.UserFileId
	isMin := req.IsMin

	// 获取文件信息
	userRpWithSavePath, err := models.FindUserReposWithSavePath(ub.UserId, userFileId, 0)
	if err != nil {
		response.RespOKFail(writer, response.FileNotExist, "文件不存在，请联系管理员")
		return
	}

	// 预览文件路径
	var previewFilePath string
	if isMin {
		// 预览最小文件
		switch userRpWithSavePath[0].FileType {
		case filehandler.IMAGE:
			previewFilePath = userRpWithSavePath[0].Path + "-pv"
		case filehandler.VIDEO:
			previewFilePath = userRpWithSavePath[0].Path + "-pv"
		default:
			response.RespOKFail(writer, response.NotSupport, "不支持该类型文件的预览")
			return
		}
	} else {
		// 预览原始文件
		previewFilePath = userRpWithSavePath[0].Path
	}
	// 打开文件
	file, err := os.OpenFile(previewFilePath, os.O_RDONLY, 0777)
	defer file.Close()
	if err != nil {
		response.RespOKFail(writer, response.FileNotExist, "预览文件不存在")
		return
	}
	// 使用io.Copy时，视频将无法拖动进度条；http.ServeContent能够解决
	// 参考 https://stackoverflow.com/questions/35667501/golang-seeking-through-a-video-serving-as-bytes
	// 拖动进度条时请求标头的Range字段会变动，因此ServeContent需要传输c.Request解析该字段
	http.ServeContent(writer, c.Request, userRpWithSavePath[0].FileName+"."+userRpWithSavePath[0].ExtendName, time.Now(), file)

	// 响应会出现http: wrote more than the declared Content-Length错误
	//response.RespOKSuccess(writer, response.Success, nil, "")
	return
}
