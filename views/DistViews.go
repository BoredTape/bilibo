package views

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegDist(r *gin.Engine, embedFs embed.FS) {
	fsDist, err := fs.Sub(embedFs, "dist")
	if err != nil {
		panic(err)
	}

	fsAssets, err := fs.Sub(fsDist, "assets")
	if err != nil {
		panic(err)
	}

	r.StaticFS("/assets", http.FS(fsAssets))
	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("/", http.FS(fsDist))
	})
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.FileFromFS("/favicon.ico", http.FS(fsDist))
	})
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS("/", http.FS(fsDist))
	})
}
