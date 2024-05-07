package main

import (
	"bilibo/bobo"
	"bilibo/config"
	"bilibo/log"
	"bilibo/models"
	"bilibo/scheduler"
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
	bobo.Init()
}

func main() {
	conf := config.GetConfig()
	os.RemoveAll(filepath.Join(conf.Download.Path, ".tmp"))
	os.MkdirAll(filepath.Join(conf.Download.Path, ".tmp"), os.ModePerm)

	b := bobo.GetBoBo()
	scheduler.BiliBoSched(b)
	scheduler.Start()

	web.Run()
}
