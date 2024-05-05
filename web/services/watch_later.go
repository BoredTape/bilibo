package services

import (
	"bilibo/consts"
	"bilibo/models"
)

func GetWatchLaterInfoByMid(mid int) *models.WatchLater {
	db := models.GetDB()
	var info models.WatchLater
	db.Where("mid = ?", mid).First(&info)
	return &info
}

func SetWatchLaterSync(mid int, sync int) {
	db := models.GetDB()
	var info models.WatchLater
	db.Where("mid = ?", mid).First(&info)
	if info.ID == 0 {
		info.Mid = mid
	}
	info.Sync = sync
	db.Save(&info)
	sql := db.Model(&models.Videos{}).Where(
		"mid = ? AND type = ?", mid, consts.VIDEO_TYPE_WATCH_LATER,
	)
	if sync == consts.WATCH_LATER_NEED_SYNC {
		sql.Where("status NOT IN (?)",
			[]int{
				consts.VIDEO_STATUS_DOWNLOAD_DONE,
				consts.VIDEO_STATUS_DOWNLOAD_FAIL,
				consts.VIDEO_STATUS_DOWNLOAD_RETRY,
				consts.VIDEO_STATUS_TO_BE_DOWNLOAD,
			})
		sql.Update("status", consts.VIDEO_STATUS_TO_BE_DOWNLOAD)
	} else if sync == consts.WATCH_LATER_NOT_SYNC {
		sql.Where(
			"status = ?", consts.VIDEO_STATUS_TO_BE_DOWNLOAD,
		)
		sql.Update("status", consts.VIDEO_STATUS_INIT)
	}
	db.Save(&info)
}
