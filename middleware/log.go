package middleware

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	common "netdisk_in_go/common/logger"
	"netdisk_in_go/models"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	// 同步write
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func (w bodyLogWriter) WriteString(s string) (int, error) {
	// 同步WriteString
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func NetdiskLogger(c *gin.Context) {
	// https://blog.csdn.net/qq_39272466/article/details/131291889
	bodyLogWriter1 := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}

	c.Writer = bodyLogWriter1
	//开始时间
	startTime := time.Now()
	//处理请求
	c.Next()
	// 异步记录
	go func(c *gin.Context) {

		responseBody := bodyLogWriter1.body.String()
		endTime := time.Now()
		// 记录格式
		ubtmp, boo := c.Get("userBasic")
		if !boo {
			return
		}
		ub := ubtmp.(*models.UserBasic)
		common.NetDiskLogger.Info("========================== start ==========================")
		common.NetDiskLogger.Info(fmt.Sprintf("user_id：%s", ub.UserId))
		common.NetDiskLogger.Info(fmt.Sprintf("user_name： %s", ub.Username))
		common.NetDiskLogger.Info(fmt.Sprintf("req time：%s", startTime.Local().Format("2006-01-02 15:04:05")))
		common.NetDiskLogger.Info(fmt.Sprintf("req url：%s", c.Request.RequestURI))
		common.NetDiskLogger.Info(fmt.Sprintf("req method：%s", c.Request.Method))
		common.NetDiskLogger.Info(fmt.Sprintf("ip：%s", c.ClientIP()))
		//common.NetDiskLogger.Info(fmt.Sprintf("query：%s", JoinParamsStr(c)))
		common.NetDiskLogger.Info(fmt.Sprintf("response body in json：%s", responseBody))
		common.NetDiskLogger.Info(fmt.Sprintf("spend time：%d ms", endTime.Sub(startTime).Milliseconds()))
		common.NetDiskLogger.Info("========================== start ==========================")
	}(c.Copy())
}
