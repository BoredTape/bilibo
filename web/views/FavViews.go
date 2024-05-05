package views

import (
	"bilibo/consts"
	"bilibo/web/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegFav(rg *gin.RouterGroup) {
	fav := rg.Group("fav")
	fav.GET("account_fav", accountFav)
	fav.POST("set_sync", SetSyncFav)
	fav.POST("set_not_sync", SetNotSyncFav)
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
