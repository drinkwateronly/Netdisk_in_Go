package share_service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"strings"
	"time"
)

// FilesShare
// @Summary 分享文件
// @Accept json
// @Produce json
// @Param req body api_models.FileShareReq true "请求"
// @Success 200 {object} api_models.RespDataList{datalist=[]api_models.RecoveryListRespAPI} "服务器响应成功，根据响应code判断是否成功"
// @Failure 400 {object} string "参数出错"
// @Router /share/sharefile [POST]
func FilesShare(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	// 绑定请求参数
	var req api.FileShareReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespBadReq(writer, "参数错误1")
		return
	}

	// 处理参数
	endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
	if err != nil {
		response.RespOK(writer, response.ReqParamNotValid, false, nil, "时间格式错误")
		return
	}
	if !time.Now().Before(endTime) {
		response.RespOK(writer, response.ShareExpired, false, nil, "分享文件已过期")
		return
	}
	shareFileIds := strings.Split(req.UserFileIds, ",") // 所有分享的用户文件id

	// 获取分享的用户文件记录
	userRps, isExist := models.FindUserFileByIds(ub.UserId, shareFileIds)
	if !isExist {
		response.RespOK(writer, 9999, false, nil, "文件缺失")
		return
	}

	// 新增分享文件记录
	shareBatchId := common.GenerateUUID()
	extractionCode := "" // 默认没有验证码
	if req.ShareType == 1 {
		// 验证码
		extractionCode = common.GenerateRandCode()
	}
	salt := common.MakeSalt() // 用于数据库存储分享验证码时加盐

	// 根据分享的用户文件记录，查询所有相关的文件记录
	// 若分享文件中包含某个文件夹，则需要查询出该文件夹内的所有文件
	// 版本1：程序递归
	var recursive func(userRps *[]models.UserRepository, userId, curPath, shareBatchId string) []models.ShareRepository
	recursive = func(userRps *[]models.UserRepository, userId, curPath, shareBatchId string) []models.ShareRepository {
		var shareRps []models.ShareRepository
		for _, userRp := range *userRps {
			shareRps = append(shareRps,
				models.ShareRepository{
					UserFileId:    userRp.UserFileId,
					ShareBatchId:  shareBatchId,
					ShareFilePath: curPath,
					FileName:      userRp.FileName,
					ExtendName:    userRp.ExtendName,
					FileSize:      userRp.FileSize,
					FileType:      userRp.FileType,
					IsDir:         userRp.IsDir,
				})
			// 如果是文件夹
			if userRp.IsDir == 1 {
				var nextUserRps []models.UserRepository
				var nextPath string
				models.DB.Where("user_id = ? and parent_id = ?", userId, userRp.UserFileId).Find(&nextUserRps)
				if curPath == "/" {
					nextPath = "/" + userRp.FileName
				} else {
					nextPath = curPath + "/" + userRp.FileName
				}
				nextShareRps := recursive(&nextUserRps, userId, nextPath, shareBatchId)
				shareRps = append(shareRps, nextShareRps...)
			}
		}
		return shareRps
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 版本1：程序递归，递归过程处理分享文件路径
		//shareRps := recursive(userRps, ub.UserId, "/", shareBatchId)

		// 版本2：sql递归，递归结束处理分享文件路径
		var shareRps []models.ShareRepository
		for _, userRp := range *userRps {
			if userRp.IsDir == 0 {
				// 不是文件夹，文件则直接放到分享文件根目录中
				shareRps = append(shareRps, models.ShareRepository{
					UserFileId:    userRp.UserFileId,
					ShareBatchId:  shareBatchId,
					ShareFilePath: "/",
					FileName:      userRp.FileName,
					ExtendName:    userRp.ExtendName,
					FileSize:      userRp.FileSize,
					FileType:      userRp.FileType,
					IsDir:         userRp.IsDir,
				})
				continue
			}
			// 是文件夹，则递归查询文件夹下的所有文件
			filesInFolder := make([]models.UserRepository, 0)
			err = models.DB.Raw(`with RECURSIVE temp as
(
    SELECT * from user_repository where user_file_id=?
    UNION ALL
    SELECT ur.* from user_repository as ur,temp t 
	where ur.parent_id=t.user_file_id AND ur.deleted_at is NULL
)
select * from temp;`, userRp.UserFileId).Find(&filesInFolder).Error
			// 循环文件夹下的文件记录，
			// 如果文件夹userRp为"/123/456"，其FilePath为"/123" -> "/"
			// 其内部文件的FilePath为"/123/456" -> "/456"
			// 分享文件不可以"/"开始
			removeLen := len(userRp.FilePath)
			for _, fileInFolder := range filesInFolder {
				shareFilePath := "/" + fileInFolder.FilePath[removeLen:]
				shareRps = append(shareRps, models.ShareRepository{
					UserFileId:    fileInFolder.UserFileId,
					ShareBatchId:  shareBatchId,
					ShareFilePath: shareFilePath,
					FileName:      fileInFolder.FileName,
					ExtendName:    fileInFolder.ExtendName,
					FileSize:      fileInFolder.FileSize,
					FileType:      fileInFolder.FileType,
					IsDir:         fileInFolder.IsDir,
				})
			}
		}

		err := tx.Create(&shareRps).Error
		if err != nil {
			return err
		}
		err = tx.Create(&models.ShareBasic{
			UserId:         ub.UserId,
			Salt:           salt,
			ShareBatchId:   shareBatchId,
			ShareType:      req.ShareType,
			ExtractionCode: common.MakePassword(shareBatchId, salt),
			ExpireTime:     endTime,
		}).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		response.RespOK(writer, response.DATABASEERROR, false, nil, err.Error())
		return
	}
	response.RespOK(writer, 0, true, gin.H{
		"shareBatchNum":  shareBatchId,
		"extractionCode": extractionCode,
	}, "")
	return
}

// CheckShareEndTime
// @Summary 检查分享文件是否过期
// @Accept json
// @Produce json
// @Param shareBatchNum query string true "分享批次id"
// @Success 200 {object} api_models.RespData{} "服务器响应成功，根据响应code判断是否成功"
// @Router /share/checkextractioncode [GET]
func CheckShareEndTime(c *gin.Context) {
	writer := c.Writer
	shareBatchId := c.Query("shareBatchNum")
	shareBasic := models.ShareBasic{}
	err := models.DB.Where("share_batch_id = ?", shareBatchId).First(&shareBasic).Error
	if err != nil {
		response.RespOK(writer, 99999, false, nil, "分享记录不存在")
		return
	}
	if shareBasic.ExpireTime.Before(time.Now()) {
		response.RespOK(writer, 99999, false, nil, "分享已过期")
		return
	}
	response.RespOK(writer, 0, true, nil, "分享有效")
}

// CheckShareType
// @Summary 检查文件分享类型
// @Accept json
// @Produce json
// @Param shareBatchNum query string true "分享批次id"
// @Success 200 {object} api_models.RespData{data=api_models.CheckShareTypeResp} "服务器响应成功，根据响应code判断是否成功"
// @Router /share/sharetype [GET]
func CheckShareType(c *gin.Context) {
	writer := c.Writer
	shareBatchId := c.Query("shareBatchNum")
	shareBasic := models.ShareBasic{}
	err := models.DB.Where("share_batch_id = ?", shareBatchId).First(&shareBasic).Error
	if err != nil {
		response.RespOK(writer, response.DatabaseError, false, nil, "分享记录不存在")
		return
	}
	response.RespOK(writer, response.Success, true, api.CheckShareTypeResp{ShareType: shareBasic.ShareType}, "分享类型")
}

// CheckShareExtractionCode
// @Summary 校验分享提取码
// @Accept json
// @Produce json
// @Param shareBatchNum query string true "分享批次id"
// @Success 200 {object} api_models.RespData{} "服务器响应成功，根据响应code判断是否成功"
// @Router /share/checkextractioncode [GET]
func CheckShareExtractionCode(c *gin.Context) {
	writer := c.Writer
	shareBatchId := c.Query("shareBatchNum")
	extractionCode := c.Query("extractionCode")

	shareBasic := models.ShareBasic{}
	err := models.DB.Where("share_batch_id = ?", shareBatchId).First(&shareBasic)
	if err != nil {
		response.RespOK(writer, response.ShareExpired, false, nil, "分享批次不存在或已过期")
		return
	}

	if common.ValidatePassword(extractionCode, shareBasic.Salt, shareBasic.ExtractionCode) {
		response.RespOK(writer, response.Success, true, nil, "验证成功")
		return
	}
	response.RespOK(writer, response.ExtractionCodeNotValid, false, nil, "提取码出错")
}

// GetShareFileList
// @Summary 获取请求路径下的分享文件列表
// @Accept json
// @Produce json
// @Param req query api_models.GetShareFileListReq true "请求"
// @Success 200 {object} api_models.RespDataList{dataList=api_models.GetShareFileListResp} "服务器响应成功，根据响应code判断是否成功"
// @Router /share/sharefileList [GET]
func GetShareFileList(c *gin.Context) {
	writer := c.Writer
	// 绑定请求参数
	req := api.GetShareFileListReq{}
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespOK(writer, response.ReqParamNotValid, false, nil, "请求参数非法")
		return
	}
	// 根据请求路径查询分享文件列表
	var shareFiles []api.GetShareFileListResp
	err = models.DB.Model(models.ShareRepository{}).
		Where("share_batch_id = ? and share_file_path = ?", req.ShareBatchId, req.ShareFilePath).Scan(&shareFiles).Error
	if err != nil {
		response.RespOK(writer, response.FileRecordNotExist, false, nil, "分享文件不存在")
		return
	}
	response.RespOkWithDataList(writer, response.Success, shareFiles, len(shareFiles), "分享文件列表")
}

// SaveShareFile
// @Summary 保存分享文件
// @Accept json
// @Produce json
// @Param req query api_models.SaveShareReq true "请求"
// @Success 200 {object} api_models.RespDataList{dataList=api_models.GetShareFileListResp} "服务器响应成功，根据响应code判断是否成功"
// @Router /share/savesharefile [POST]
func SaveShareFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	var req api.SaveShareReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespOK(writer, response.ReqParamNotValid, false, nil, "请求参数非法")
		return
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 查询父文件夹是否存在
		parentDir, isExist, err := models.FindParentDirFromAbsPath(tx, ub.UserId, req.FilePath)
		if err != nil {
			return err
		}
		if !isExist {
			return errors.New("file not found")
		}

		// 获得所有的分享文件的id
		userFileIds := strings.Split(req.UserFileIds, ",")
		if len(userFileIds) <= 0 {
			return errors.New("share file not found")
		}

		// 查询分享文件记录
		var userRps []models.UserRepository
		err = models.DB.Where("user_file_id in ?", userFileIds).Find(&userRps).Error
		if err != nil {
			response.RespOK(writer, 9999, false, nil, err.Error())
			return nil
		}

		// 循环获取所有分享文件的新记录，若某个文件是文件夹，则进入文件夹获取记录
		var recursive func(tx *gorm.DB, curFiles []models.UserRepository, parentId, curPath, userId string) *[]models.UserRepository
		recursive = func(tx *gorm.DB, curFiles []models.UserRepository, parentId, curPath, userId string) *[]models.UserRepository {
			var urs []models.UserRepository
			for _, curFile := range curFiles {
				newRecord := models.UserRepository{
					UserFileId:    common.GenerateUUID(),
					FileId:        curFile.FileId,
					UserId:        userId,
					FilePath:      curPath,
					ParentId:      parentId,
					FileName:      curFile.FileName,
					ExtendName:    curFile.ExtendName,
					FileType:      curFile.FileType,
					IsDir:         curFile.IsDir,
					FileSize:      curFile.FileSize,
					ModifyTime:    time.Now().Format("2006-01-02 15:04:05"),
					UploadTime:    time.Now().Format("2006-01-02 15:04:05"),
					DeleteBatchId: "",
				}
				urs = append(urs, newRecord)
				if curFile.IsDir == 1 { // 是文件夹
					var nextFile []models.UserRepository
					tx.Where("parent_id = ? and user_id = ?", curFile.UserFileId, userId).Find(&nextFile)
					var nextPath string
					if curPath == "/" {
						nextPath = "/" + curFile.FileName
					} else {
						nextPath = curPath + "/" + curFile.FileName
					}
					tmp := recursive(tx, nextFile, curFile.UserFileId, nextPath, userId)
					urs = append(urs, *tmp...)
				}
			}
			return &urs
		}

		newShareFiles := recursive(tx, userRps, parentDir.UserFileId, req.FilePath, ub.UserId)
		if len(*newShareFiles) == 0 {
			return errors.New("share file not found")
		}
		if err := tx.Create(newShareFiles).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		response.RespOK(writer, 9999, false, nil, err.Error())
		return
	}
	response.RespOK(writer, 0, true, nil, "上传成功")
	return

}

// GetShareList
// @Summary 获取用户的分享记录
// @Accept json
// @Produce json
// @Param req query api_models.GetShareListReq true "请求"
// @Success 200 {object} api_models.RespDataList{dataList=api_models.GetShareFileListResp} "服务器响应成功，根据响应code判断是否成功"
// @Router /share/sharefileList [GET]
func GetShareList(c *gin.Context) {
	writer := c.Writer
	ub := c.MustGet("userBasic").(*models.UserBasic)
	var req api.GetShareListReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespOK(writer, response.ReqParamNotValid, false, nil, "请求参数非法")
		return
	}
	if req.ShareFilePath != "/" {
		response.RespOK(writer, response.NotSupport, false, nil, "暂不支持进入分享文件夹查看")
		return
	}
	files, total, err := models.FindShareFilesByPathAndPage(ub.UserId, req)
	response.RespOkWithDataList(writer, response.ReqParamNotValid, files, total, "分享文件列表")
}
