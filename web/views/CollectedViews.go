package views

import (
	"bilibo/consts"
	"bilibo/web/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegCollected(rg *gin.RouterGroup) {
	collected := rg.Group("collected")
	collected.GET("account_collected", accountCollected)
	collected.POST("set_sync", SetSyncCollected)
	collected.POST("set_not_sync", SetNotSyncCollected)
}

func accountCollected(c *gin.Context) {
	rsp := gin.H{
		"data":    nil,
		"message": "",
		"result":  0,
	}

	if midStr, ok := c.GetQuery("mid"); ok {
		if mid, err := strconv.Atoi(midStr); err == nil {
			datas := services.GetAccountCollectIdInfoByMid(mid)
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

type CollectedSetStatus struct {
	Mid    int `json:"mid" binding:"required"`
	CollId int `json:"coll_id" binding:"required"`
}

func SetSyncCollected(c *gin.Context) {
	var req CollectedSetStatus
	c.BindJSON(&req)
	services.SetCollectedSyncStatus(req.Mid, req.CollId, consts.COLLECTED_NEED_SYNC)
	rsp := gin.H{
		"message": "",
		"result":  0,
	}
	c.JSON(http.StatusOK, rsp)
}

func SetNotSyncCollected(c *gin.Context) {
	var req CollectedSetStatus
	c.BindJSON(&req)
	services.SetCollectedSyncStatus(req.Mid, req.CollId, consts.FAVOUR_NOT_SYNC)
	rsp := gin.H{
		"message": "",
		"result":  0,
	}
	c.JSON(http.StatusOK, rsp)
}
