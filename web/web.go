package web

import (
	"bilibo/config"
	"bilibo/log"
	"fmt"

	"github.com/gin-gonic/gin"
)

func Run() {
	logger := log.GetLogger()
	conf := config.GetConfig()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	Route(r)

	for _, route := range r.Routes() {
		logger.Infof("%s [%s]", route.Path, route.Method)
	}
	logger.Infof("web server running on %s:%d", conf.Server.Host, conf.Server.Port)
	if err := r.Run(fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port)); err != nil {
		panic(err)
	}
}
