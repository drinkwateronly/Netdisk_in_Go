package share_service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"netdisk_in_go/common"
	"netdisk_in_go/common/api"
	"netdisk_in_go/common/filehandler"
	"netdisk_in_go/common/response"
	"netdisk_in_go/models"
	"strings"
	"time"
)

// FilesShare
// @Summary 分享文件
// @Description 生成分享文件链接与分享提取码，设置分享过期时间
// @Tags share
// @Accept json
// @Produce json
// @Param req body api.FileShareReq true "请求"
// @Success 200 {object} response.RespData{data=api.FileShareResp} "响应"
// @Router /share/sharefile [POST]
func FilesShare(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub, boo := models.GetUserBasicFromContext(c)
	if !boo {
		response.RespUnAuthorized(writer)
	}
	// 绑定请求参数
	var req api.FileShareReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespBadReq(writer, "参数错误")
		return
	}

	// 处理参数
	endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime) // 分享过期时间
	if err != nil {
		response.RespOK(writer, response.ReqParamNotValid, false, nil, "时间格式错误")
		return
	}
	if !time.Now().Before(endTime) {
		response.RespOKFail(writer, response.ShareExpired, "过期时间不能早于当前时间")
		return
	}

	shareFileIds := strings.Split(req.UserFileIds, ",") // 所有分享的用户文件id

	// 获取分享的用户文件记录
	userRps, err := models.FindUserFilesByIds(models.DB, ub.UserId, shareFileIds)
	if err != nil {
		response.RespOKFail(writer, response.ShareFileNotExist, "文件缺失")
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
	// 程序递归查询函数
	var recursive func(tx *gorm.DB, userRps *[]models.UserRepository, userId, curPath, shareBatchId string) []models.ShareRepository
	recursive = func(tx *gorm.DB, userRps *[]models.UserRepository, userId, curPath, shareBatchId string) []models.ShareRepository {
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
				tx.Where("user_id = ? and parent_id = ?", userId, userRp.UserFileId).Find(&nextUserRps)
				if curPath == "/" {
					nextPath = "/" + userRp.FileName
				} else {
					nextPath = curPath + "/" + userRp.FileName
				}
				nextShareRps := recursive(tx, &nextUserRps, userId, nextPath, shareBatchId)
				shareRps = append(shareRps, nextShareRps...)
			}
		}
		return shareRps
	}

	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 版本1：程序递归，递归过程处理分享文件路径
		//shareRps := recursive(tx, userRps, ub.UserId, "/", shareBatchId)

		// 版本2：sql递归，递归结束处理分享文件路径
		var shareRps []models.ShareRepository
		for _, userRp := range userRps {
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
			// 分享的是文件夹，则递归查询文件夹下的所有文件
			filesInFolder := make([]models.UserRepository, 0)
			err = tx.Raw(`with RECURSIVE temp as
(
    SELECT * from user_repository where user_file_id = ? AND deleted_at = 0
    UNION ALL
    SELECT ur.* from user_repository as ur,temp t 
	where ur.parent_id=t.user_file_id AND ur.deleted_at = 0
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
					ParentId:      fileInFolder.ParentId,
					ShareBatchId:  shareBatchId,
					ShareFilePath: shareFilePath,
					FileId:        fileInFolder.FileId,
					FileName:      fileInFolder.FileName,
					ExtendName:    fileInFolder.ExtendName,
					FileSize:      fileInFolder.FileSize,
					FileType:      fileInFolder.FileType,
					IsDir:         fileInFolder.IsDir,
				})
			}
		}
		// 新增分享文件库记录
		err := tx.Create(&shareRps).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "生成分享失败")
			return err
		}
		// 新增分享批次记录
		err = tx.Create(&models.ShareBasic{
			UserId:         ub.UserId,
			Salt:           salt,
			ShareBatchId:   shareBatchId,
			ShareType:      req.ShareType,
			ExtractionCode: common.MakePassword(extractionCode, salt),
			ExpireTime:     endTime,
		}).Error
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "生成分享失败")
			return err
		}

		response.RespOKSuccess(writer, response.Success, api.FileShareResp{
			ShareBatchId:   shareBatchId,
			ExtractionCode: extractionCode,
		}, "分享已生成")
		return nil
	})
}

// CheckShareEndTime
// @Summary 检查分享文件是否过期
// @Description
// @Tags share
// @Accept json
// @Produce json
// @Param req query api.CheckShareReq true "请求"
// @Success 200 {object} response.RespData{} "响应"
// @Router /share/checkendtime [GET]
func CheckShareEndTime(c *gin.Context) {
	writer := c.Writer
	// 绑定参数
	var req api.CheckShareReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
		return
	}
	// 查询分享批次记录
	shareBasic := models.ShareBasic{}
	err = models.DB.Where("share_batch_id = ?", req.ShareBatchId).First(&shareBasic).Error
	if err != nil {
		response.RespOKFail(writer, response.ShareFileNotExist, "分享记录不存在")
		return
	}
	// 检查时间
	if shareBasic.ExpireTime.Before(time.Now()) {
		response.RespOKFail(writer, response.ShareExpired, "分享已过期")
		return
	}
	response.RespOKSuccess(writer, response.Success, nil, "分享有效")
}

// CheckShareType
// @Summary 检查文件分享类型
// @Description
// @Tags share
// @Accept json
// @Produce json
// @Param req query api.CheckShareReq true "req"
// @Success 200 {object} response.RespData{data=api.CheckShareTypeResp} "resp"
// @Router /share/sharetype [GET]
func CheckShareType(c *gin.Context) {
	writer := c.Writer
	// 绑定请求参数
	var req api.CheckShareReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
		return
	}
	// 查询分享批次记录
	shareBasic := models.ShareBasic{}
	err = models.DB.Where("share_batch_id = ?", req.ShareBatchId).First(&shareBasic).Error
	if err != nil {
		response.RespOKFail(writer, response.DatabaseError, "分享记录不存在")
		return
	}
	//
	response.RespOKSuccess(writer, response.Success, api.CheckShareTypeResp{
		ShareType: shareBasic.ShareType,
	}, "分享类型")
}

// CheckShareExtractionCode
// @Summary 校验分享提取码
// @Description
// @Tags share
// @Accept json
// @Produce json
// @Param shareBatchNum query api.CheckExtractionCodeReq true "req"
// @Success 200 {object} response.RespData{} "resp"
// @Router /share/checkextractioncode [GET]
func CheckShareExtractionCode(c *gin.Context) {
	writer := c.Writer
	// 绑定请求参数
	var req api.CheckExtractionCodeReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数错误")
		return
	}
	// 查询分享批次记录
	shareBasic := models.ShareBasic{}
	err = models.DB.Where("share_batch_id = ?", req.ShareBatchId).First(&shareBasic).Error
	if err != nil {
		response.RespOK(writer, response.ShareExpired, false, nil, "分享批次不存在或已过期")
		return
	}

	if !common.ValidatePassword(req.ExtractionCode, shareBasic.Salt, shareBasic.ExtractionCode) {
		response.RespOKFail(writer, response.ExtractionCodeNotValid, "提取码不正确")
		return
	}

	response.RespOKSuccess(writer, response.Success, nil, "提取码验证成功")
	return
}

// GetShareFileList
// @Summary 获取请求路径下的分享文件列表
// @Description
// @Tags share
// @Accept json
// @Produce json
// @Param req query api.GetShareFileListReq true "请求"
// @Success 200 {object} response.RespDataList{dataList=api.GetShareFileListResp} "响应"
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
// @Description
// @Tags share
// @Accept json
// @Produce json
// @Param req body api.SaveShareReq true "请求"
// @Success 200 {object} response.RespDataList{dataList=api.GetShareFileListResp} "服务器响应成功，根据响应code判断是否成功"
// @Router /share/savesharefile [POST]
func SaveShareFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub, boo := models.GetUserBasicFromContext(c)
	if !boo {
		response.RespUnAuthorized(writer)
		return
	}
	// 绑定请求参数
	var req api.SaveShareReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.RespBadReq(writer, "请求参数非法")
		return
	}
	// 获得所有的分享文件的id
	userFileIds := strings.Split(req.UserFileIds, ",")
	if len(userFileIds) == 0 {
		response.RespOKFail(writer, response.ShareFileNotExist, "获取失败")
		return
	}
	// 开启事务
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		// 查询存放分享文件的目标文件夹是否存在
		destDir, err := models.FindParentDirFromFilePath(tx, ub.UserId, req.FilePath)
		if err != nil {
			response.RespOK(writer, response.FileNotExist, false, nil, "目标文件夹不存在")
			return err
		}
		// 存放要新增的用户文件记录
		newFiles := make([]models.UserRepository, 0, len(userFileIds))
		for _, userFileId := range userFileIds {
			// 如果是文件夹，则找到文件夹下所有文件记录；如果是文件，则找到文件记录。
			curShareFiles, err := models.FindAllShareFilesFromFileId(tx, req.ShareBatchNum, userFileId)
			// 没找到文件
			if err != nil {
				response.RespOKFail(writer, response.ShareFileNotExist, "分享文件夹不存在")
				return err
			}
			// 文件名
			prePathLen := len(curShareFiles[0].ShareFilePath)
			newPath := req.FilePath
			fileIdMap := make(map[string]string)
			for i := range curShareFiles {
				var ur models.UserRepository
				if i == 0 {
					// 保存分享的只有一个文件时，获取该分享
					newUUID := common.GenerateUUID()
					fileIdMap[curShareFiles[0].UserFileId] = newUUID
					ur = models.UserRepository{
						UserFileId: newUUID,                 // 新用户文件id
						FileId:     curShareFiles[0].FileId, // 文件实际保存地址
						UserId:     ub.UserId,               // 文件所有者
						FilePath:   newPath,                 // 路径
						ParentId:   destDir.UserFileId,      // 父文件夹id
						FileName:   curShareFiles[0].FileName,
						ExtendName: curShareFiles[0].ExtendName,
						FileType:   curShareFiles[0].FileType,
						IsDir:      curShareFiles[0].IsDir,
						FileSize:   curShareFiles[0].FileSize,
						UploadTime: time.Now().Format("2006-01-02 15:04:05"),
					}
				} else {
					newUUID := common.GenerateUUID() // 文件新uuid
					newParentId, ok := fileIdMap[curShareFiles[i].ParentId]
					if !ok {
						response.RespOKFail(writer, response.DatabaseError, err.Error())
						return errors.New("parent not found")
					}
					ur = models.UserRepository{
						UserFileId:    newUUID,
						FileId:        curShareFiles[i].FileId,
						UserId:        ub.UserId,
						FilePath:      filehandler.ConCatFileFullPath(newPath, curShareFiles[i].ShareFilePath[prePathLen:]),
						ParentId:      newParentId,
						FileName:      curShareFiles[i].FileName,
						ExtendName:    curShareFiles[i].ExtendName,
						FileType:      curShareFiles[i].FileType,
						IsDir:         curShareFiles[i].IsDir,
						FileSize:      curShareFiles[i].FileSize,
						ModifyTime:    "",
						UploadTime:    time.Now().Format("2006-01-02 15:04:05"),
						DeleteBatchId: "",
					}
					// 如果当前文件是文件夹，则记录该文件夹id的新旧映射
					if curShareFiles[i].IsDir == 1 {
						fileIdMap[curShareFiles[i].UserFileId] = newUUID
					}
				}
				newFiles = append(newFiles, ur)
			}
		}
		err = tx.Create(newFiles).Error
		if models.IsDuplicateEntryErr(err) {
			response.RespOKFail(writer, response.FileRepeat, "有重复文件，请更换文件存放位置")
			return err
		}
		if err != nil {
			response.RespOKFail(writer, response.DatabaseError, "有重复文件，请更换文件存放位置")
			return err
		}
		response.RespOKSuccess(writer, response.Success, nil, "上传成功")
		return nil
	})

}

// GetMyShareList
// @Summary 获取用户的分享记录
// @Description 根据分享批次和路径获取用户自己的已分享文件列表
// @Tags share
// @Accept json
// @Produce json
// @Param req query api.GetShareListReq true "请求"
// @Success 200 {object} response.RespDataList{dataList=api.GetShareFileListResp} "响应"
// @Router /share/shareList [GET]
func GetMyShareList(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub, boo := models.GetUserBasicFromContext(c)
	if !boo {
		response.RespUnAuthorized(writer)
		return
	}
	// 绑定参数
	var req api.GetShareListReq
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.RespOK(writer, response.ReqParamNotValid, false, nil, "请求参数非法")
		return
	}
	// 查找
	files, total, err := models.FindMyShareList(ub.UserId, req)
	if err != nil && !models.IsRecordNotFoundErr(err) {
		// 允许文件记录不存在
		response.RespOKFail(writer, response.DatabaseError, "获取失败")
		return
	}
	response.RespOkWithDataList(writer, response.Success, files, total, "分享文件列表")
}
