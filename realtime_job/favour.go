package realtime_job

import (
	"bilibo/config"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/models"
	"bilibo/utils"
	"fmt"
	"strings"
	"time"
)

func ChangeFavourName(mlid int, oldPath, newPath string) {
	db := models.GetDB()
	logger := log.GetLogger()
	sqlPause := "UPDATE favour_videos SET status=status+100 WHERE status IN (?) AND deleted_at IS NULL;"
	value := []int{
		consts.VIDEO_STATUS_TO_BE_DOWNLOAD,
		consts.VIDEO_STATUS_DOWNLOAD_FAIL,
		consts.VIDEO_STATUS_DOWNLOAD_RETRY,
	}
	db.Exec(sqlPause, value)
	for {
		logger.Info(fmt.Sprintf("收藏夹路径更改:\n%s => %s", oldPath, newPath))
		var downloadingCount int64
		db.Model(&models.FavourVideos{}).Where(
			"mlid = ? AND status = ?", mlid, consts.VIDEO_STATUS_DOWNLOADING,
		).Count(&downloadingCount)
		fmt.Println(downloadingCount)
		if downloadingCount == 0 {
			fmt.Println(oldPath, "\n", newPath)
			if err := utils.RenameDir(oldPath, newPath); err != nil {
				logger.Error(err.Error())
			}
			sqlContinue := "UPDATE favour_videos SET status=status-100 WHERE status > 100 AND deleted_at IS NULL;"
			db.Exec(sqlContinue)
			break
		} else {
			logger.Info(fmt.Sprintf("收藏夹路径 %s 正在下载,重试中...", oldPath))
		}
		time.Sleep(2 * time.Second)
	}
}

func DeleteFavours(mlids []int) {
	db := models.GetDB()
	logger := log.GetLogger()
	favInfos := []models.FavourFoldersInfo{}
	db.Where("mlid IN (?)", mlids).Find(&favInfos)
	conf := config.GetConfig()
	basePath := conf.Download.Path
	db.Where(
		"mlid IN (?) AND status != ?", mlids, consts.VIDEO_STATUS_DOWNLOADING,
	).Delete(&models.FavourVideos{})
	for {
		logger.Info(fmt.Sprintf("删除收藏夹,收藏夹IDs:[%s]", strings.Trim(strings.Replace(fmt.Sprint(mlids), " ", ",", -1), "[]")))
		var downloadingCount int64
		db.Model(&models.FavourVideos{}).Where(
			"mlid IN (?) AND status = ?", mlids, consts.VIDEO_STATUS_DOWNLOADING,
		).Count(&downloadingCount)
		if downloadingCount == 0 {
			db.Where("mlid IN (?)", mlids).Delete(&models.FavourFoldersInfo{})
			db.Where("mlid IN (?)", mlids).Delete(&models.FavourVideos{})
			for _, fav := range favInfos {
				utils.RecyclePath(fav.Mid, basePath, utils.Name(fav.Title))
			}
			break
		} else {
			logger.Info(fmt.Sprintf("收藏夹视频正在下载,重试中..."))
		}
		time.Sleep(2 * time.Second)
	}
}
