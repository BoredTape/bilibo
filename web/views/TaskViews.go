package views

import (
	"bilibo/web/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegTask(rg *gin.RouterGroup) {
	task := rg.Group("task")
	task.GET("list", taskList)
}

func taskList(c *gin.Context) {
	rsp := gin.H{
		"data":    nil,
		"message": "",
		"result":  0,
	}
	rsp["data"] = services.TaskList()

	c.JSON(http.StatusOK, rsp)
}
