package services

import (
	"bilibo/consts"
	"bilibo/log"
	"bilibo/models"
	"fmt"
	"time"

	"golang.org/x/exp/maps"
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
		favourVideo.Status = consts.VIDEO_STATUS_TO_BE_DOWNLOAD
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

func DelFavourVideoByMid(mid int) {
	db := models.GetDB()
	db.Where(models.FavourVideos{Mid: mid}).Delete(&models.FavourVideos{})
}

type VideoInfo struct {
	Part        string `json:"part"`
	Title       string `json:"title"`
	Bvid        string `json:"bvid"`
	Status      int    `json:"status"`
	Mlid        int    `json:"mlid"`
	FavTitle    string `json:"fav_title"`
	Mid         int    `json:"mid"`
	AccountName string `json:"account_name"`
}

func GetVideosByStatus(status, page, pageSize int) (*[]*VideoInfo, int64) {
	result := make([]*VideoInfo, 0)
	db := models.GetDB()
	var total int64

	statusList := []int{status}

	var mlids []int
	if status == consts.VIDEO_STATUS_TO_BE_DOWNLOAD {
		var syncFav []models.FavourFoldersInfo
		db.Model(&models.FavourFoldersInfo{}).Where(&models.FavourFoldersInfo{Sync: consts.FAVOUR_NEED_SYNC}).Find(&syncFav)
		for _, v := range syncFav {
			mlids = append(mlids, v.Mlid)
		}
	}

	if status == consts.VIDEO_STATUS_DOWNLOAD_FAIL {
		statusList = append(statusList, consts.VIDEO_STATUS_DOWNLOAD_RETRY)
	}

	if len(mlids) > 0 && status == consts.VIDEO_STATUS_TO_BE_DOWNLOAD {
		db.Model(&models.FavourVideos{}).Where("mlid IN (?) AND status = ?", mlids, status).Count(&total)
	} else if len(mlids) < 1 && status == consts.VIDEO_STATUS_TO_BE_DOWNLOAD {
		return &result, 0
	} else {
		db.Model(&models.FavourVideos{}).Where("status IN (?)", statusList).Count(&total)
	}

	if total > 0 {
		var favourVideos []models.FavourVideos
		if status == consts.VIDEO_STATUS_TO_BE_DOWNLOAD {
			db.Model(&models.FavourVideos{}).Where("mlid IN (?) AND status = ?", mlids, status).Limit(pageSize).Offset((page - 1) * pageSize).Find(&favourVideos)
		} else {
			db.Model(&models.FavourVideos{}).Where("status IN (?)", statusList).Limit(pageSize).Offset((page - 1) * pageSize).Find(&favourVideos)
		}
		accountMap := make(map[int]*AccountInfo, 0)
		favMap := make(map[int]*FavourFolders, 0)
		for _, v := range favourVideos {
			accountMap[v.Mid] = nil
			favMap[v.Mlid] = nil
		}

		var favourFolderInfos []models.FavourFoldersInfo
		db.Where("mid IN (?)", maps.Keys(accountMap)).Find(&favourFolderInfos)
		for _, v := range favourFolderInfos {
			favMap[v.Mlid] = &FavourFolders{
				Mlid:       v.Mlid,
				Fid:        v.Fid,
				Title:      v.Title,
				MediaCount: v.MediaCount,
				Sync:       v.Sync,
			}
		}

		var accountInfos []models.BiliAccounts
		db.Where("mid IN (?)", maps.Keys(accountMap)).Find(&accountInfos)
		for _, v := range accountInfos {
			accountMap[v.Mid] = &AccountInfo{
				Mid:    v.Mid,
				Status: v.Status,
				Face:   v.Face,
				Uname:  v.UName,
			}
		}

		for _, v := range favourVideos {
			favTitle := ""
			if favMap[v.Mlid] != nil {
				favTitle = favMap[v.Mlid].Title
			}
			accountName := ""
			if accountMap[v.Mid] != nil {
				accountName = accountMap[v.Mid].Uname
			}
			result = append(result, &VideoInfo{
				Part:        fmt.Sprintf("P%d %s", v.Page, v.Part),
				Title:       v.Title,
				Bvid:        v.Bvid,
				Status:      v.Status,
				Mlid:        v.Mlid,
				FavTitle:    favTitle,
				Mid:         v.Mid,
				AccountName: accountName,
			})
		}
	}

	return &result, total
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
