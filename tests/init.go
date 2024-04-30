package tests

import (
	"bilibo/config"
	"bilibo/log"
	"bilibo/models"
)

func InitConfig() {
	c := config.Config{
		Server: config.ServerConfig{
			DB: config.DBConfig{
				Driver: "sqlite",
				DSN:    "./data.db",
			},
			Host: "127.0.0.1",
			Port: 8080,
		},
		Download: config.DownloadConfig{
			Path: "./downloads",
		},
	}
	config.SetConfig(&c)
}

func Init() {
	InitConfig()
	log.InitLogger()
	conf := config.GetConfig()
	models.InitDB(conf.Server.DB.Driver, conf.Server.DB.DSN)
}
