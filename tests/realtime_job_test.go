package tests

import (
	"bilibo/config"
	"bilibo/consts"
	"bilibo/models"
	"bilibo/services"
	"bilibo/utils"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setup() {
	os.RemoveAll(config.GetConfig().Download.Path)
	favPath := utils.GetFavourPath(1, config.GetConfig().Download.Path)
	recyclePath := utils.GetRecyclePath(1, config.GetConfig().Download.Path)
	os.MkdirAll(filepath.Join(favPath, "test"), os.ModePerm)
	os.MkdirAll(recyclePath, os.ModePerm)

	db := models.GetDB()
	db.Migrator().DropTable(
		&models.BiliAccounts{},
		&models.FavourFoldersInfo{},
		&models.Videos{},
	)
	db.AutoMigrate(
		&models.BiliAccounts{},
		&models.FavourFoldersInfo{},
		&models.Videos{},
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

	video1 := models.Videos{
		SourceId:       1,
		Mid:            1,
		Bvid:           "abc1",
		Cid:            1,
		Status:         consts.VIDEO_STATUS_DOWNLOADING,
		LastDownloadAt: nil,
		Type:           consts.VIDEO_TYPE_FAVOUR,
	}
	db.Save(&video1)

	video2 := models.Videos{
		SourceId:       1,
		Mid:            1,
		Bvid:           "abc2",
		Cid:            1,
		Status:         consts.VIDEO_STATUS_DOWNLOAD_FAIL,
		LastDownloadAt: nil,
		Type:           consts.VIDEO_TYPE_FAVOUR,
	}
	db.Save(&video2)

	video3 := models.Videos{
		SourceId:       1,
		Mid:            1,
		Bvid:           "abc3",
		Cid:            1,
		Status:         consts.VIDEO_STATUS_DOWNLOAD_RETRY,
		LastDownloadAt: nil,
		Type:           consts.VIDEO_TYPE_FAVOUR,
	}
	db.Save(&video3)

	video4 := models.Videos{
		SourceId:       1,
		Mid:            1,
		Bvid:           "abc4",
		Cid:            1,
		Status:         consts.VIDEO_STATUS_TO_BE_DOWNLOAD,
		LastDownloadAt: nil,
		Type:           consts.VIDEO_TYPE_FAVOUR,
	}
	db.Save(&video4)
}

func teardown() {
	db := models.GetDB()
	db.Migrator().DropTable(
		&models.BiliAccounts{},
		&models.FavourFoldersInfo{},
		&models.Videos{},
	)
	os.RemoveAll(config.GetConfig().Download.Path)
}

func ChangeFavourName(t *testing.T) {
	setup()
	defer teardown()

	db := models.GetDB()

	var downloadingCount1 int64
	db.Model(&models.Videos{}).Where(
		"status IN (?)", []int{consts.VIDEO_STATUS_DOWNLOAD_FAIL, consts.VIDEO_STATUS_DOWNLOAD_RETRY, consts.VIDEO_STATUS_TO_BE_DOWNLOAD},
	).Count(&downloadingCount1)

	if downloadingCount1 != 3 {
		t.Fatal("downloadingCount != 3")
	}
	basePath := utils.GetFavourPath(1, config.GetConfig().Download.Path)
	go services.ChangeFavourName(1,
		filepath.Join(basePath, "test"),
		filepath.Join(basePath, "test1"),
	)

	time.Sleep(5 * time.Second)

	var downloadingCount2 int64
	db.Model(&models.Videos{}).Where(
		"status < 100",
	).Count(&downloadingCount2)

	if downloadingCount2 != 1 {
		t.Log(downloadingCount2)
		t.Fatal("downloadingCount != 1")
	}
	time.Sleep(2 * time.Second)
	db.Model(&models.Videos{}).Where(
		"status = ?", consts.VIDEO_STATUS_DOWNLOADING,
	).Update("status", consts.VIDEO_STATUS_DOWNLOAD_DONE)
	time.Sleep(3 * time.Second)

	favPath := utils.GetFavourPath(1, config.GetConfig().Download.Path)
	_, err := os.Stat(filepath.Join(favPath, "test1"))
	if err != nil {
		t.Fatal(err)
	}
}

func DeleteFavour(t *testing.T) {
	setup()
	defer teardown()
	db := models.GetDB()
	go services.DeleteFavours(1, []int{1})

	time.Sleep(5 * time.Second)
	db.Model(&models.Videos{}).Where(
		"status = ?", consts.VIDEO_STATUS_DOWNLOADING,
	).Update("status", consts.VIDEO_STATUS_DOWNLOAD_DONE)
	time.Sleep(3 * time.Second)
	recyclePath := utils.GetRecyclePath(1, config.GetConfig().Download.Path)
	dirFiles, err := os.ReadDir(recyclePath)
	if err != nil {
		t.Fatal(err)
	}
	fileNames := make([]string, 0)

	for _, file := range dirFiles {
		f, err := file.Info()
		if err != nil {
			t.Fatal(err)
		}
		fileNames = append(fileNames, f.Name())
	}
	if len(fileNames) == 0 {
		t.Fatal("没有成功删除")
	}
}

func TestFileUtils(t *testing.T) {
	Init()
	ChangeFavourName(t)
	DeleteFavour(t)
}
