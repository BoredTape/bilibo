package views

import (
	"bilibo/config"
	"bilibo/log"
	"bilibo/utils"
	"bilibo/web/services"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

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
	account.GET("dir", AccountDir)
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
	services.DelFavourInfoByMid(req.Mid)
	services.DelVideoByMid(req.Mid)
	services.DelAccount(req.Mid)
	services.DelWatchLaterByMid(req.Mid)
	services.DelCollectedByMid(req.Mid)
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

	url, qrId, err := services.SetAccountInfo()
	if err == nil {
		data["url"] = url
		data["id"] = fmt.Sprintf("%d", qrId)
		rsp["data"] = data
		rsp["message"] = "获取登陆二维码成功"
		rsp["result"] = 0
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
	faceUrlEncode := c.Query("url")
	faceUrlDecode, err := url.QueryUnescape(faceUrlEncode)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	resp, err := http.Get(faceUrlDecode)

	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500, gin.H{"message": err.Error()})
		return
	}
	c.Writer.Write(body)
}

func AccountDir(c *gin.Context) {
	logger := log.GetLogger()
	rsp := gin.H{
		"message": "",
		"result":  0,
	}
	if queryMap, err := utils.GetQueryMap(c, []string{"q", "adapter"}); err != nil {
		rsp["result"] = 999
		rsp["message"] = err.Error()
		c.JSON(http.StatusOK, rsp)
		return
	} else {
		path := c.DefaultQuery("path", queryMap["adapter"]+"://")
		if queryMap["q"] == "index" {
			rsp = services.GetAccountIndex(queryMap["adapter"], queryMap["q"], path)
			c.JSON(http.StatusOK, rsp)
			return
		} else if (queryMap["q"] == "preview" || queryMap["q"] == "download") && path != "" {
			filePath, err := services.GetAccountFileDownload(queryMap["adapter"], queryMap["q"], path)
			if err != nil {
				logger.Error("get favour file download error: %v", err)
				rsp["result"] = 999
				rsp["message"] = err.Error()
				c.JSON(http.StatusOK, rsp)
			} else {
				logger.Info("get favour file download: %s", filePath)
				fileNameSplit := strings.Split(path, "/")
				slices.Reverse(fileNameSplit)
				fileName := fileNameSplit[0]
				c.Header("Content-Description", "Simulation File Download")
				c.Header("Content-Transfer-Encoding", "binary")
				c.Header("Content-Disposition", "attachment; filename="+fileName)
				c.Header("Content-Type", "application/octet-stream")
				c.File(filePath)
			}
		}
	}
}
