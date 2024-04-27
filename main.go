package main

import (
	"bilibo/bili"
	"bilibo/config"
	"bilibo/download"
	"bilibo/log"
	"bilibo/models"
	"bilibo/scheduler"
	"bilibo/services"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func init() {
	config.InitConfig()
	conf := config.GetConfig()
	log.InitLogger()
	models.InitDB(conf.Server.DB.Driver, conf.Server.DB.DSN)
	bili.InitBiliBo()
	services.InitSetVideoStatus()
}

func main() {
	logger := log.GetLogger()
	conf := config.GetConfig()
	os.RemoveAll(filepath.Join(conf.Download.Path, ".tmp"))
	os.MkdirAll(filepath.Join(conf.Download.Path, ".tmp"), os.ModePerm)

	bobo := bili.GetBilibo()
	scheduler.BiliBoSched(bobo)
	scheduler.Start()

	for _, clientId := range bobo.ClientList() {
		ctx, cancel := context.WithCancel(context.Background())
		go download.AccountDownload(clientId, ctx)
		bobo.ClientSetCancal(clientId, cancel)
	}
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
