package tests

import (
	"bilibo/config"
	"bilibo/consts"
	"bilibo/models"
	"bilibo/services"
	"bilibo/tests"
	"bilibo/utils"
	"os"
	"path/filepath"
	"testing"
)

func setup() {
	os.RemoveAll(config.GetConfig().Download.Path)
	favPath := utils.GetFavourPath(1, config.GetConfig().Download.Path)
	recyclePath := utils.GetRecyclePath(1, config.GetConfig().Download.Path)
	os.MkdirAll(filepath.Join(favPath, "test", "abc1"), os.ModePerm)
	os.MkdirAll(recyclePath, os.ModePerm)

	db := models.GetDB()
	db.Migrator().DropTable(
		&models.BiliAccounts{},
		&models.FavourFoldersInfo{},
		&models.Videos{},
		&models.VideosInfo{},
	)
	db.AutoMigrate(
		&models.BiliAccounts{},
		&models.FavourFoldersInfo{},
		&models.Videos{},
		&models.VideosInfo{},
	)

	account := models.BiliAccounts{
		Mid:     1,
		UName:   "test",
		Face:    "test",
		ImgKey:  "test",
		SubKey:  "test",
		Cookies: "test",
		Status:  consts.ACCOUNT_STATUS_NORMAL,
	}
	db.Save(&account)

	fav := models.FavourFoldersInfo{
		Mid:        1,
		Fid:        1,
		MediaCount: 1,
		Attr:       1,
		Title:      "test",
		Mlid:       1,
		FavState:   1,
		Sync:       1,
	}
	db.Save(&fav)

	video := models.Videos{
		SourceId:       1,
		Mid:            1,
		Bvid:           "abc1",
		Cid:            1,
		Status:         consts.VIDEO_STATUS_DOWNLOAD_DONE,
		LastDownloadAt: nil,
		Type:           consts.VIDEO_TYPE_FAVOUR,
	}
	db.Save(&video)

	videoInfo := models.VideosInfo{
		Bvid:   "abc1",
		Cid:    1,
		Page:   1,
		Title:  "abc1",
		Part:   "testvideo",
		Width:  1080,
		Height: 720,
		Rotate: 0,
	}
	db.Save(&videoInfo)
}

func SetInvalidVideos(t *testing.T) {
	services.SetInvalidVideos(1, 1, []string{"abc1"}, consts.VIDEO_TYPE_FAVOUR)
	favPath := utils.GetFavourPath(1, config.GetConfig().Download.Path)
	distPath := filepath.Join(favPath, "test", "abc1[已失效]")

	if _, err := os.Stat(distPath); err != nil {
		t.Fatal(err)
	}
	db := models.GetDB()
	var videosCount int64 = 0
	db.Model(&models.Videos{}).Where("bvid = ?", "abc1").Count(&videosCount)
	if videosCount == 0 {
		t.Fatal("videosCount == 0")
	}

	var videoInfoCount int64 = 0
	db.Model(&models.VideosInfo{}).Where("bvid = ?", "abc1").Count(&videoInfoCount)
	if videoInfoCount == 0 {
		t.Fatal("videoInfoCount == 0")
	}
}

func teardown() {
	db := models.GetDB()
	db.Migrator().DropTable(
		&models.BiliAccounts{},
		&models.FavourFoldersInfo{},
		&models.Videos{},
		&models.VideosInfo{},
	)
	os.RemoveAll(config.GetConfig().Download.Path)
}

func TestInvalidVideos(t *testing.T) {
	tests.Init()
	setup()
	defer teardown()
	SetInvalidVideos(t)
}
