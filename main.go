package main

import (
	"bilibo/bobo"
	"bilibo/config"
	"bilibo/log"
	"bilibo/models"
	"bilibo/scheduler"
	"bilibo/services"
	"bilibo/universal"
	"bilibo/web"
	"os"
	"path/filepath"
)

func init() {
	config.Init()
	universal.Init()
	log.Init()
	conf := config.GetConfig()
	models.Init(conf.Server.DB.Driver, conf.Server.DB.DSN)
	services.InitSetVideoStatus()
	bobo.Init()
}

func main() {
	conf := config.GetConfig()
	os.RemoveAll(filepath.Join(conf.Download.Path, ".tmp"))
	os.MkdirAll(filepath.Join(conf.Download.Path, ".tmp"), os.ModePerm)

	scheduler.BiliBoSched(bobo.GetBoBo())
	scheduler.Start()

	web.Run()
}
