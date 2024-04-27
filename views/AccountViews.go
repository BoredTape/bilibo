package views

import (
	"bilibo/bili"
	"bilibo/config"
	"bilibo/consts"
	"bilibo/download"
	"bilibo/log"
	"bilibo/scheduler"
	"bilibo/services"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func RegAccount(rg *gin.RouterGroup) {
	account := rg.Group("account")
	account.GET("list", accountList)
	account.POST("delete", accountDelete)
	account.GET("save", accountSave)
	account.GET("proxy/:mid", accountProxy)
	account.GET("/qrcode/:fileName", accountQrCode)
	account.GET("/qrcode/status/:id", accountQrCodeStatus)
}

func accountList(c *gin.Context) {
	data := gin.H{}
	rsp := gin.H{
		"data":    &data,
		"message": "",
		"result":  0,
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		rsp["message"] = err.Error()
		rsp["result"] = 999
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		rsp["message"] = err.Error()
		rsp["result"] = 999
	}
	if rsp["result"] != 0 {
		c.JSON(http.StatusOK, rsp)
		return
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 10 {
		pageSize = 10
	}

	data["page"] = page
	data["page_size"] = pageSize
	items, total := services.AccountList(page, pageSize)
	data["total"] = total
	data["items"] = items

	c.JSON(http.StatusOK, rsp)
}

type accountDeleteReq struct {
	Mid int `json:"mid" binding:"required"`
}

func accountDelete(c *gin.Context) {
	var req accountDeleteReq
	c.BindJSON(&req)
	bilibo := bili.GetBilibo()
	bilibo.DelClient(req.Mid)
	services.DelAccount(req.Mid)
	services.DelFavourInfoByMid(req.Mid)
	services.DelFavourVideoByMid(req.Mid)
	jobIds := services.DelTaskByMid(req.Mid)
	for _, jobId := range jobIds {
		scheduler.DelJob(jobId)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "account delete",
		"result":  0,
	})
}

func accountSave(c *gin.Context) {
	data := make(map[string]interface{})
	rsp := gin.H{
		"data":    data,
		"message": "获取登陆二维码失败",
		"result":  999,
	}
	client := bili.NewClient("", 0)

	if qr, err := client.GetQRCode(); err == nil {
		if qrImgByte, err := qr.Encode(); err == nil {
			conf := config.GetConfig()
			qrId := time.Now().UnixNano()
			fileName := fmt.Sprintf("%d.png", qrId)
			filePath := filepath.Join(conf.Download.Path, ".tmp", fileName)
			if err := os.WriteFile(filePath, qrImgByte, os.ModePerm); err == nil {
				data["url"] = "/api/account/qrcode/" + fileName
				data["id"] = fmt.Sprintf("%d", qrId)
				rsp["data"] = data
				rsp["message"] = "获取登陆二维码成功"
				rsp["result"] = 0
				services.AddQRCodeInfo(fmt.Sprintf("%d", qrId))
				go func() {
					logger := log.GetLogger()
					if err := client.LoginWithQRCode(qr); err == nil {
						biliBo := bili.GetBilibo()
						biliBo.AddClient(client)
						nav, _, err := client.GetNavigation()
						if err != nil {
							logger.Error(err)
							return
						}
						if err := client.RefreshWbiKey(nav); err != nil {
							logger.Error(err)
							return
						}
						imgKey, subKey := client.GetWbiRunningTime()
						services.SaveAccountInfo(
							client.GetMid(),
							nav.Uname, nav.Face,
							client.GetCookiesString(),
							imgKey, subKey,
						)
						ctx, cancel := context.WithCancel(context.Background())
						biliBo.ClientSetCancal(client.GetMid(), cancel)
						go download.AccountDownload(client.GetMid(), ctx)
						services.SetQRCodeStatus(fmt.Sprintf("%d", qrId), consts.QRCODE_STATUS_SCANNED)
					} else {
						logger.Error(err)
						services.SetQRCodeStatus(fmt.Sprintf("%d", qrId), consts.QRCODE_STATUS_INVALID)
					}
					os.Remove(filePath)
				}()
			}
		}
	}

	c.JSON(http.StatusOK, rsp)
}

func accountQrCode(c *gin.Context) {
	if fileName := c.Param("fileName"); fileName == "" {
		c.Status(404)
	} else {
		filePath := filepath.Join(config.GetConfig().Download.Path, ".tmp", fileName)
		c.File(filePath)
	}
}

func accountQrCodeStatus(c *gin.Context) {
	rsp := gin.H{
		"data":    nil,
		"message": "qrcode not found",
		"result":  999,
	}
	if qrId := c.Param("id"); qrId == "" {
		c.Status(404)
	} else {
		if info := services.GetQRCodeInfo(qrId); info.ID != 0 {
			data := map[string]interface{}{
				"status": info.Status,
			}
			rsp["data"] = data
			rsp["message"] = "qrcode status"
			rsp["result"] = 0
		}
	}
	c.JSON(http.StatusOK, rsp)
}

func accountProxy(c *gin.Context) {
	mid, err := strconv.Atoi(c.Param("mid"))
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	faceUrlEncode := c.Query("url")
	faceUrlDecode, err := url.QueryUnescape(faceUrlEncode)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	bilibo := bili.GetBilibo()
	client, err := bilibo.GetClient(mid)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	resty := client.GetResty()
	resp, err := resty.R().Get(faceUrlDecode)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	c.Writer.Write(resp.Body())
}
