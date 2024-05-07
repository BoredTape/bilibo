package models

import (
	"bilibo/consts"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func Init(driver, dsn string) {
	var err error
	gConfig := gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
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

	db.Migrator().DropTable(&Task{}, &QRCode{})
	db.AutoMigrate(
		&BiliAccounts{},
		&FavourFoldersInfo{},
		&Videos{},
		&VideosInfo{},
		&QRCode{},
		&VideoDownloadMessage{},
		&Task{},
		&WatchLater{},
		&CollectedInfo{},
	)
	db.Model(&Videos{}).Where("status = ?", consts.VIDEO_STATUS_DOWNLOADING).Update("status", consts.VIDEO_STATUS_TO_BE_DOWNLOAD)
}

func GetDB() *gorm.DB {
	return db
}
