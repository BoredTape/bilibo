package web

import (
	"bilibo/web/views"
	"embed"

	"github.com/gin-gonic/gin"
)

//go:embed dist
var embedFs embed.FS

func Route(r *gin.Engine) {
	views.RegDist(r, embedFs)
	api := r.Group("api")
	views.RegAccount(api)
	views.RegFav(api)
	views.RegVideo(api)
	views.RegWatchLater(api)
	views.RegTask(api)
}
