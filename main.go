package main

import (
	"bilibo/bili"
	"bilibo/config"
	"bilibo/download"
	"bilibo/log"
	"bilibo/models"
	"bilibo/scheduler"
	"bilibo/services"
	"bilibo/web"
	"context"
	"os"
	"path/filepath"
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
	web.Run()
}
