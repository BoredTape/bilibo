package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func Init(driver, dsn string) {
	var err error
	gConfig := gorm.Config{Logger: logger.Default.LogMode(logger.Info)}
	if driver == "mysql" {
		db, err = gorm.Open(mysql.Open(dsn), &gConfig)
		if err != nil {
			panic(err)
		}
	} else if driver == "sqlite" {
		db, err = gorm.Open(sqlite.Open(dsn), &gConfig)
		if err != nil {
			panic(err)
		}
	} else {
		panic("数据库驱动不支持")
	}

	db.AutoMigrate(
		&BiliAccounts{},
		&FavourFoldersInfo{},
		&FavourVideos{},
		&QRCode{},
		&VideoDownloadMessage{},
	)
}

func GetDB() *gorm.DB {
	return db
}
