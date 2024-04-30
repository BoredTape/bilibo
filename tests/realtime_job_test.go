package tests

import (
	"bilibo/config"
	"bilibo/consts"
	"bilibo/models"
	"bilibo/realtime_job"
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
		&models.FavourVideos{},
	)
	db.AutoMigrate(
		&models.BiliAccounts{},
		&models.FavourFoldersInfo{},
		&models.FavourVideos{},
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

	video1 := models.FavourVideos{
		Mlid:           1,
		Mid:            1,
		Bvid:           "abc1",
		Cid:            1,
		Page:           1,
		Title:          "testVideo2",
		Part:           "testPart2",
		Width:          1,
		Height:         1,
		Rotate:         1,
		Status:         consts.VIDEO_STATUS_DOWNLOADING,
		LastDownloadAt: nil,
	}
	db.Save(&video1)

	video2 := models.FavourVideos{
		Mlid:           1,
		Mid:            1,
		Bvid:           "abc2",
		Cid:            1,
		Page:           1,
		Title:          "testVideo2",
		Part:           "testPart2",
		Width:          1,
		Height:         1,
		Rotate:         1,
		Status:         consts.VIDEO_STATUS_DOWNLOAD_FAIL,
		LastDownloadAt: nil,
	}
	db.Save(&video2)

	video3 := models.FavourVideos{
		Mlid:           1,
		Mid:            1,
		Bvid:           "abc3",
		Cid:            1,
		Page:           1,
		Title:          "testVideo3",
		Part:           "testPart3",
		Width:          1,
		Height:         1,
		Rotate:         1,
		Status:         consts.VIDEO_STATUS_DOWNLOAD_RETRY,
		LastDownloadAt: nil,
	}
	db.Save(&video3)

	video4 := models.FavourVideos{
		Mlid:           1,
		Mid:            1,
		Bvid:           "abc4",
		Cid:            1,
		Page:           1,
		Title:          "testVideo4",
		Part:           "testPart4",
		Width:          1,
		Height:         1,
		Rotate:         1,
		Status:         consts.VIDEO_STATUS_TO_BE_DOWNLOAD,
		LastDownloadAt: nil,
	}
	db.Save(&video4)
}

func teardown() {
	db := models.GetDB()
	db.Migrator().DropTable(
		&models.BiliAccounts{},
		&models.FavourFoldersInfo{},
		&models.FavourVideos{},
	)
	os.RemoveAll(config.GetConfig().Download.Path)
}

func ChangeFavourName(t *testing.T) {
	setup()
	defer teardown()

	db := models.GetDB()

	var downloadingCount1 int64
	db.Model(&models.FavourVideos{}).Where(
		"status IN (?)", []int{consts.VIDEO_STATUS_DOWNLOAD_FAIL, consts.VIDEO_STATUS_DOWNLOAD_RETRY, consts.VIDEO_STATUS_TO_BE_DOWNLOAD},
	).Count(&downloadingCount1)

	if downloadingCount1 != 3 {
		t.Fatal("downloadingCount != 3")
	}
	basePath := utils.GetFavourPath(1, config.GetConfig().Download.Path)
	go realtime_job.ChangeFavourName(1,
		filepath.Join(basePath, "test"),
		filepath.Join(basePath, "test1"),
	)

	time.Sleep(5 * time.Second)

	var downloadingCount2 int64
	db.Model(&models.FavourVideos{}).Where(
		"status < 100",
	).Count(&downloadingCount2)
	if downloadingCount2 != 1 {
		t.Log(downloadingCount2)
		t.Fatal("downloadingCount != 0")
	}
	time.Sleep(2 * time.Second)
	db.Model(&models.FavourVideos{}).Where(
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
	go realtime_job.DeleteFavours([]int{1})

	time.Sleep(5 * time.Second)
	db.Model(&models.FavourVideos{}).Where(
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
