package views

import (
	"bilibo/consts"
	"bilibo/log"
	"bilibo/utils"
	"bilibo/web/services"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegFav(rg *gin.RouterGroup) {
	fav := rg.Group("fav")
	fav.GET("account_fav", accountFav)
	fav.POST("set_sync", SetSyncFav)
	fav.POST("set_not_sync", SetNotSyncFav)
	fav.GET("dir", FavDir)
}

func accountFav(c *gin.Context) {
	rsp := gin.H{
		"data":    nil,
		"message": "",
		"result":  0,
	}

	if midStr, ok := c.GetQuery("mid"); ok {
		if mid, err := strconv.Atoi(midStr); err == nil {
			datas := services.GetAccountFavourInfoByMid(mid)
			rsp["data"] = datas
		} else {
			rsp["message"] = err.Error()
			rsp["result"] = 999
		}
	} else {
		rsp["message"] = "mid not found"
		rsp["result"] = 999
	}

	c.JSON(http.StatusOK, rsp)
}

type FavSetStatus struct {
	Mid  int `json:"mid" binding:"required"`
	Mlid int `json:"mlid" binding:"required"`
}

func SetSyncFav(c *gin.Context) {
	var req FavSetStatus
	c.BindJSON(&req)
	services.SetFavourSyncStatus(req.Mid, req.Mlid, consts.FAVOUR_NEED_SYNC)
	rsp := gin.H{
		"message": "",
		"result":  0,
	}
	c.JSON(http.StatusOK, rsp)
}

func SetNotSyncFav(c *gin.Context) {
	var req FavSetStatus
	c.BindJSON(&req)
	services.SetFavourSyncStatus(req.Mid, req.Mlid, consts.FAVOUR_NOT_SYNC)
	rsp := gin.H{
		"message": "",
		"result":  0,
	}
	c.JSON(http.StatusOK, rsp)
}

func FavDir(c *gin.Context) {
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
			rsp = services.GetFavourIndex(queryMap["adapter"], queryMap["q"], path)
			c.JSON(http.StatusOK, rsp)
			return
		} else if (queryMap["q"] == "preview" || queryMap["q"] == "download") && path != "" {
			filePath, err := services.GetFavourFileDownload(queryMap["adapter"], queryMap["q"], path)
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
