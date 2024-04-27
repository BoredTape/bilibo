package views

import (
	"bilibo/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RegVideo(rg *gin.RouterGroup) {
	video := rg.Group("video")
	video.GET("list", VideoList)
}

func VideoList(c *gin.Context) {
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

	status, err := strconv.Atoi(c.DefaultQuery("status", "0"))
	if err != nil {
		rsp["message"] = err.Error()
		rsp["result"] = 999
	}
	items, total := services.GetVideosByStatus(status, page, pageSize)
	data["page"] = page
	data["page_size"] = pageSize
	data["total"] = total
	data["items"] = items

	c.JSON(http.StatusOK, rsp)
}
