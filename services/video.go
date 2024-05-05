package services

import (
	"bilibo/consts"
	"bilibo/log"
	"bilibo/models"
	"time"
)

type VideoService struct {
	V *models.Videos
}

func (f *VideoService) SetMid(mid int) {
	f.V = &models.Videos{
		Mid: mid,
	}
}

func (f *VideoService) Save() {
	db := models.GetDB()
	var video models.Videos
	db.Where(models.Videos{
		Bvid: f.V.Bvid,
		Mlid: f.V.Mlid,
		Mid:  f.V.Mid,
		Cid:  f.V.Cid,
		Type: f.V.Type,
	}).FirstOrInit(&video)
	needUpdata := false
	if video.ID == 0 && f.V.Type == consts.VIDEO_TYPE_WATCH_LATER {
		video.Status = consts.VIDEO_STATUS_INIT
		needUpdata = true
	} else if video.ID == 0 && f.V.Type == consts.VIDEO_TYPE_FAVOUR {
		favInfo := GetFavourInfoByMlid(f.V.Mlid)
		video.Status = consts.VIDEO_STATUS_INIT
		if favInfo != nil && favInfo.Sync == consts.FAVOUR_NEED_SYNC {
			video.Status = consts.VIDEO_STATUS_TO_BE_DOWNLOAD
		}
		needUpdata = true
	}
	if video.Part != f.V.Part {
		video.Part = f.V.Part
		needUpdata = true
	}
	if video.Width != f.V.Width {
		video.Width = f.V.Width
		needUpdata = true
	}
	if video.Height != f.V.Height {
		video.Height = f.V.Height
		needUpdata = true
	}
	if video.Rotate != f.V.Rotate {
		video.Rotate = f.V.Rotate
		needUpdata = true
	}
	if video.Title != f.V.Title {
		video.Title = f.V.Title
		needUpdata = true
	}
	if video.Page != f.V.Page {
		video.Page = f.V.Page
		needUpdata = true
	}
	if needUpdata {
		db.Save(&video)
	}

}

func GetVideoByMidStatus(mid, status int) *models.Videos {
	db := models.GetDB()
	var video models.Videos
	subQuery := db.Model(&models.FavourFoldersInfo{}).Where(
		&models.FavourFoldersInfo{Mid: mid, Sync: consts.FAVOUR_NEED_SYNC},
	).Select("mlid")
	db.Where(
		"mid = ? AND status = ? AND (mlid IN (?) OR mlid = 0)",
		mid, status, subQuery,
	).First(&video)
	if video.ID == 0 {
		return nil
	}
	return &video
}

func GetToBeDownloadByMid(mid int) *models.Videos {
	return GetVideoByMidStatus(mid, consts.VIDEO_STATUS_TO_BE_DOWNLOAD)
}

func GetRetryByMid(mid int) *models.Videos {
	before, _ := time.ParseDuration("-2h")
	last_time := time.Now().Add(before)
	db := models.GetDB()
	var video models.Videos
	subQuery := db.Model(&models.FavourFoldersInfo{}).Where(
		&models.FavourFoldersInfo{Mid: mid, Sync: consts.FAVOUR_NEED_SYNC},
	).Select("mlid")
	db.Model(&models.Videos{}).Where(
		"mid = ? AND status = ? AND (mlid IN (?) OR mlid = 0) AND last_download_at < ?",
		mid, consts.VIDEO_STATUS_DOWNLOAD_RETRY, subQuery, last_time,
	).First(&video)
	if video.ID == 0 {
		return nil
	}
	return &video
}

func SetVideoStatus(id uint, status int) {
	db := models.GetDB()
	var video models.Videos
	db.Where("id = ?", id).First(&video)
	video.Status = status
	if video.Status == consts.VIDEO_STATUS_DOWNLOADING {
		timeNow := time.Now()
		video.LastDownloadAt = &timeNow
	}
	db.Save(&video)
}

func SetUserVideosStatus(mid int, status int) {
	db := models.GetDB()
	db.Model(&models.Videos{}).Where(
		"mid = ? AND status != ?", mid, status,
	).Updates(map[string]interface{}{"status": status})
}

func SetVideoErrorMessage(Mlid, Mid int, Bvid, Error string) {
	logger := log.GetLogger()
	logger.Errorf(Error)
	db := models.GetDB()
	errorInfo := models.VideoDownloadMessage{
		Mlid:    Mlid,
		Mid:     Mid,
		Bvid:    Bvid,
		Message: Error,
		Type:    consts.VIDEO_MESSAGE_ERROR,
	}
	db.Create(&errorInfo)
}
