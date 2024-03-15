package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	ApiModels "netdisk_in_go/api_models"
	"netdisk_in_go/models"
	"netdisk_in_go/utils"
	"strings"
	"time"
)

func ShareFiles(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	var req ApiModels.FileShareReq
	err := c.ShouldBind(&req)
	if err != nil {
		utils.RespBadReq(writer, "参数错误1")
		return
	}

	// 处理参数
	endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
	if err != nil || !time.Now().Before(endTime) { // 过期时间在当前时间之前
		utils.RespOK(writer, 9999, false, nil, "时间格式错误或过期")
		return
	}
	fileIds := strings.Split(req.UserFileIds, ",")

	// 获取分享的用户文件记录
	userRps, isExist := models.FindUserFileByIds(ub.UserId, fileIds)
	if !isExist {
		utils.RespOK(writer, 9999, false, nil, "文件缺失")
		return
	}

	// 新增分享文件记录
	shareBatchId := utils.GenerateUUID()
	var extractionCode string
	if req.ShareType == 1 {
		extractionCode = utils.GenerateRandCode()
	}
	salt := utils.MakeSalt()

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
				utils.DB.Where("user_id = ? and parent_id = ?", userId, userRp.UserFileId).Find(&nextUserRps)
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

	err = utils.DB.Transaction(func(tx *gorm.DB) error {
		shareRps := recursive(userRps, ub.UserId, "/", shareBatchId)
		err := tx.Create(&shareRps).Error
		if err != nil {
			return err
		}
		err = tx.Create(&models.ShareBasic{
			Salt:           salt,
			ShareBatchId:   shareBatchId,
			ShareType:      req.ShareType,
			ExtractionCode: utils.MakePassword(shareBatchId, salt),
			ExpireTime:     endTime,
		}).Error
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.RespOK(writer, ApiModels.DATABASEERROR, false, nil, err.Error())
		return
	}
	utils.RespOK(writer, 0, true, gin.H{
		"shareBatchNum":  shareBatchId,
		"extractionCode": extractionCode,
	}, "")
	return
}

func CheckShareEndTime(c *gin.Context) {
	writer := c.Writer
	shareBatchId := c.Query("shareBatchNum")
	shareBasic := models.ShareBasic{}
	res := utils.DB.Where("share_batch_id = ?", shareBatchId).Find(&shareBasic)
	if res.RowsAffected == 0 {
		utils.RespOK(writer, 99999, false, nil, "分享记录不存在")
		return
	}
	if shareBasic.ExpireTime.Before(time.Now()) {
		utils.RespOK(writer, 99999, false, nil, "分享已过期")
		return
	}
	utils.RespOK(writer, 0, true, nil, "分享有效")
}

func CheckShareType(c *gin.Context) {
	writer := c.Writer
	shareBatchId := c.Query("shareBatchNum")
	shareBasic := models.ShareBasic{}
	_ = utils.DB.Where("share_batch_id = ?", shareBatchId).Find(&shareBasic)
	utils.RespOK(writer, 0, true, gin.H{"shareType": shareBasic.ShareType}, "分享类型")
}

func CheckShareExtractionCode(c *gin.Context) {
	writer := c.Writer
	shareBatchId := c.Query("shareBatchNum")
	extractionCode := c.Query("extractionCode")

	shareBasic := models.ShareBasic{}
	_ = utils.DB.Where("share_batch_id = ?", shareBatchId).Find(&shareBasic)
	fmt.Println(extractionCode, shareBatchId)

	if utils.ValidatePassword(extractionCode, shareBasic.Salt, shareBasic.ExtractionCode) {
		utils.RespOK(writer, 0, true, nil, "验证成功")
		return
	}
	utils.RespOK(writer, 9999, true, nil, "提取码出错")
}

func GetShareFileList(c *gin.Context) {
	writer := c.Writer
	// 校验cookie，获取用户信息
	shareBatchId := c.Query("shareBatchNum")
	shareFilePath := c.Query("shareFilePath")
	shareRps, total, err := models.FindShareFilesByPath(shareFilePath, shareBatchId)
	if err != nil {
		return
	}
	utils.RespOkWithDataList(writer, 0, shareRps, total, "分享文件列表")
}

func SaveShareFile(c *gin.Context) {
	writer := c.Writer
	// 获取用户信息
	ub := c.MustGet("userBasic").(*models.UserBasic)
	var req ApiModels.SaveShareReq
	err := c.ShouldBind(&req)
	if err != nil {
		utils.RespBadReq(writer, "参数错误1")
		return
	}

	err = utils.DB.Transaction(func(tx *gorm.DB) error {

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
		err = utils.DB.Where("user_file_id in ?", userFileIds).Find(&userRps).Error
		if err != nil {
			utils.RespOK(writer, 9999, false, nil, err.Error())
			return nil
		}

		// 循环获取所有分享文件的新记录，若某个文件是文件夹，则进入文件夹获取记录
		var recursive func(tx *gorm.DB, curFiles []models.UserRepository, parentId, curPath, userId string) *[]models.UserRepository
		recursive = func(tx *gorm.DB, curFiles []models.UserRepository, parentId, curPath, userId string) *[]models.UserRepository {
			var urs []models.UserRepository
			for _, curFile := range curFiles {
				newRecord := models.UserRepository{
					UserFileId:    utils.GenerateUUID(),
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
		utils.RespOK(writer, 9999, false, nil, err.Error())
		return
	}
	utils.RespOK(writer, 0, true, nil, "上传成功")
	return

}
