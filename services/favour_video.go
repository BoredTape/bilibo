package services

import (
	"bilibo/consts"
	"bilibo/log"
	"bilibo/models"
	"time"
)

type FavourVideoService struct {
	V *models.FavourVideos
}

func (f *FavourVideoService) SetMid(mid int) {
	f.V = &models.FavourVideos{
		Mid: mid,
	}
}

func (f *FavourVideoService) Save() {
	db := models.GetDB()
	var favourVideo models.FavourVideos
	db.Where(models.FavourVideos{
		Bvid: f.V.Bvid,
		Mlid: f.V.Mlid,
		Mid:  f.V.Mid,
		Cid:  f.V.Cid,
	}).FirstOrInit(&favourVideo)
	needUpdata := false
	if favourVideo.ID == 0 {
		favInfo := GetFavourInfoByMlid(f.V.Mlid)
		favourVideo.Status = consts.VIDEO_STATUS_INIT
		if favInfo != nil && favInfo.Sync == consts.FAVOUR_NEED_SYNC {
			favourVideo.Status = consts.VIDEO_STATUS_TO_BE_DOWNLOAD
		}
		needUpdata = true
	}
	if favourVideo.Part != f.V.Part {
		favourVideo.Part = f.V.Part
		needUpdata = true
	}
	if favourVideo.Width != f.V.Width {
		favourVideo.Width = f.V.Width
		needUpdata = true
	}
	if favourVideo.Height != f.V.Height {
		favourVideo.Height = f.V.Height
		needUpdata = true
	}
	if favourVideo.Rotate != f.V.Rotate {
		favourVideo.Rotate = f.V.Rotate
		needUpdata = true
	}
	if favourVideo.Title != f.V.Title {
		favourVideo.Title = f.V.Title
		needUpdata = true
	}
	if favourVideo.Page != f.V.Page {
		favourVideo.Page = f.V.Page
		needUpdata = true
	}
	if needUpdata {
		db.Save(&favourVideo)
	}

}

func GetVideoByMidStatus(mid, status int) *models.FavourVideos {
	db := models.GetDB()
	var favourVideo models.FavourVideos
	subQuery := db.Model(&models.FavourFoldersInfo{}).Where(
		&models.FavourFoldersInfo{Mid: mid, Sync: consts.FAVOUR_NEED_SYNC},
	).Select("mlid")
	db.Where(
		"mid = ? AND status = ? AND mlid IN (?)",
		mid, status, subQuery,
	).First(&favourVideo)
	if favourVideo.ID == 0 {
		return nil
	}
	return &favourVideo
}

func GetToBeDownloadByMid(mid int) *models.FavourVideos {
	return GetVideoByMidStatus(mid, consts.VIDEO_STATUS_TO_BE_DOWNLOAD)
}

func GetRetryByMid(mid int) *models.FavourVideos {
	before, _ := time.ParseDuration("-2h")
	last_time := time.Now().Add(before)
	db := models.GetDB()
	var favourVideo models.FavourVideos
	subQuery := db.Model(&models.FavourFoldersInfo{}).Where(
		&models.FavourFoldersInfo{Mid: mid, Sync: consts.FAVOUR_NEED_SYNC},
	).Select("mlid")
	db.Model(&models.FavourVideos{}).Where(
		"mid = ? AND status = ? AND mlid IN (?) AND last_download_at < ?",
		mid, consts.VIDEO_STATUS_DOWNLOAD_RETRY, subQuery, last_time,
	).First(&favourVideo)
	if favourVideo.ID == 0 {
		return nil
	}
	return &favourVideo
}

func SetVideoStatus(id uint, status int) {
	db := models.GetDB()
	var favourVideo models.FavourVideos
	db.Where("id = ?", id).First(&favourVideo)
	favourVideo.Status = status
	if favourVideo.Status == consts.VIDEO_STATUS_DOWNLOADING {
		timeNow := time.Now()
		favourVideo.LastDownloadAt = &timeNow
	}
	db.Save(&favourVideo)
}

func SetUserVideosStatus(mid int, status int) {
	db := models.GetDB()
	db.Model(&models.FavourVideos{}).Where(
		"mid = ? AND status != ?", mid, status,
	).Updates(map[string]interface{}{"status": status})
}

func InitSetVideoStatus() {
	db := models.GetDB()
	var favourVideo []models.FavourVideos
	db.Where("status = ?", consts.VIDEO_STATUS_DOWNLOADING).Find(&favourVideo)
	for _, v := range favourVideo {
		v.Status = consts.VIDEO_STATUS_TO_BE_DOWNLOAD
		db.Save(&v)
	}
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
