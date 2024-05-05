package views

import (
	"bilibo/consts"
	"bilibo/web/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegWatchLater(rg *gin.RouterGroup) {
	wl := rg.Group("watch_later")
	wl.GET("status", GetWLStatus)
	wl.POST("set_sync", SetWLSync)
	wl.POST("set_not_sync", SetWLNotSync)
}

func GetWLStatus(c *gin.Context) {
	resp := gin.H{
		"data":    nil,
		"message": "",
		"result":  0,
	}
	midStr := c.DefaultQuery("mid", "")
	if midStr != "" {
		if mid, err := strconv.Atoi(midStr); err == nil {
			resp["data"] = services.GetWatchLaterInfoByMid(mid)
		}
	}
	c.JSON(http.StatusOK, resp)
}

type SetWLStatusReq struct {
	Mid int `json:"mid" binding:"required"`
}

func SetWLSync(c *gin.Context) {
	var req SetWLStatusReq
	c.BindJSON(&req)
	services.SetWatchLaterSync(req.Mid, consts.WATCH_LATER_NEED_SYNC)
	rsp := gin.H{
		"message": "",
		"result":  0,
	}
	c.JSON(http.StatusOK, rsp)
}

func SetWLNotSync(c *gin.Context) {
	var req SetWLStatusReq
	c.BindJSON(&req)
	services.SetWatchLaterSync(req.Mid, consts.WATCH_LATER_NOT_SYNC)
	rsp := gin.H{
		"message": "",
		"result":  0,
	}
	c.JSON(http.StatusOK, rsp)
}
