package services

import (
	"bilibo/config"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/models"
	"bilibo/utils"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/exp/maps"
)

func GetToBeDownloadByMid(mid int) *models.Videos {
	db := models.GetDB()
	var video models.Videos
	db.Model(&models.Videos{}).Where(
		"mid = ? AND status = ?",
		mid, consts.VIDEO_STATUS_TO_BE_DOWNLOAD,
	).First(&video)
	if video.ID == 0 {
		return nil
	}
	return &video
}

func GetRetryByMid(mid int) *models.Videos {
	before, _ := time.ParseDuration("-2h")
	last_time := time.Now().Add(before)
	db := models.GetDB()
	var video models.Videos
	subQuery := db.Model(&models.FavourFoldersInfo{}).Where(
		&models.FavourFoldersInfo{Mid: mid, Sync: consts.FAVOUR_NEED_SYNC},
	).Select("mlid")
	subQueryCollected := db.Model(&models.CollectedInfo{}).Where(
		&models.CollectedInfo{Mid: mid, Sync: consts.COLLECTED_NEED_SYNC},
	).Select("coll_id")
	db.Model(&models.Videos{}).Where(
		"mid = ? AND status = ? AND last_download_at < ? AND ((source_id IN (?) AND type = ?) OR (type = ?)  OR (source_id IN (?) AND type = ?))",
		mid, consts.VIDEO_STATUS_DOWNLOAD_RETRY, last_time, subQuery, consts.VIDEO_TYPE_FAVOUR, consts.VIDEO_TYPE_WATCH_LATER, subQueryCollected, consts.VIDEO_TYPE_COLLECTED,
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

func SetVideoErrorMessage(SourceId, Mid, VideoType int, Bvid, Error string) {
	logger := log.GetLogger()
	logger.Errorf(Error)
	db := models.GetDB()
	errorInfo := models.VideoDownloadMessage{
		SourceId:  SourceId,
		Mid:       Mid,
		VideoType: VideoType,
		Bvid:      Bvid,
		Message:   Error,
		Type:      consts.VIDEO_MESSAGE_ERROR,
	}
	db.Create(&errorInfo)
}

type Video struct {
	SourceId int
	Mid      int
	Bvid     string
	Cid      int
	Type     int
}

type VideoInfo struct {
	Bvid   string
	Cid    int
	Page   int
	Title  string
	Part   string
	Width  int
	Height int
	Rotate int
}

func SetVideos(mid, source_id int, videos map[string]*Video, vType int) {
	db := models.GetDB()
	videosList := []models.Videos{}
	db.Model(&models.Videos{}).Where(
		&models.Videos{Mid: mid, Type: vType, SourceId: source_id},
	).Find(&videosList)

	existsVideoMap := make(map[string]models.Videos)
	for _, v := range videosList {
		mapKey := fmt.Sprintf("%d_%s_%d", v.SourceId, v.Bvid, v.Cid)
		existsVideoMap[mapKey] = v

	}
	existsKeys := maps.Keys(existsVideoMap)
	newKeys := maps.Keys(videos)

	insertKeys := utils.Difference(newKeys, existsKeys)
	deleteKeys := utils.Difference(existsKeys, newKeys)

	deleteIds := make([]uint, 0)
	if len(deleteKeys) > 0 {
		for _, key := range deleteKeys {
			if video, ok := existsVideoMap[key]; ok {
				deleteIds = append(deleteIds, video.ID)
			}
		}
		if len(deleteIds) > 0 {
			times := int(math.Ceil(float64(len(deleteIds)) / 100))
			for i := 0; i < times; i++ {
				start := i * 100
				end := i*100 + 100
				if end > len(deleteIds) {
					end = len(deleteIds)
				}
				db.Delete(&models.Videos{}, deleteIds[start:end])
			}
		}
	}

	if len(insertKeys) > 0 {
		createList := []models.Videos{}
		for _, key := range insertKeys {
			if video, ok := videos[key]; ok {
				createList = append(createList, models.Videos{
					SourceId: video.SourceId,
					Mid:      mid,
					Bvid:     video.Bvid,
					Cid:      video.Cid,
					Type:     vType,
					Status:   consts.VIDEO_STATUS_INIT,
				})
			}
		}
		if len(createList) > 0 {
			db.CreateInBatches(createList, 100)
		}
	}
}

type GroupVideo struct {
	SourceId int
	Mid      int
	Bvid     string
	Status   int
	Type     int
}

type GroupVideoInfo struct {
	Bvid  string
	Title string
}

func SetInvalidVideos(mid, source_id int, bvids []string, vType int) {
	logger := log.GetLogger()
	db := models.GetDB()
	videosList := []GroupVideo{}

	if len(bvids) > 0 {
		times := int(math.Ceil(float64(len(bvids)) / 100))
		for i := 0; i < times; i++ {
			start := i * 100
			end := i*100 + 100
			if end > len(bvids) {
				end = len(bvids)
			}
			bvidsSlice := bvids[start:end]

			rangeVideos := []GroupVideo{}

			db.Model(&models.Videos{}).Select(
				"mid", "type", "source_id", "status", "bvid",
			).Where(&models.Videos{
				Mid: mid, Type: vType, SourceId: source_id, Status: consts.VIDEO_STATUS_DOWNLOAD_DONE,
			}).Where("bvid IN (?)", bvidsSlice).
				Group("mid").
				Group("type").
				Group("source_id").
				Group("status").
				Group("bvid").
				Find(&rangeVideos)

			if len(rangeVideos) > 0 {
				videosList = append(videosList, rangeVideos...)
			}
		}
	}

	if len(videosList) < 1 {
		return
	}

	videoInfoList := []GroupVideoInfo{}
	if len(bvids) > 0 {
		times := int(math.Ceil(float64(len(bvids)) / 100))
		for i := 0; i < times; i++ {
			start := i * 100
			end := i*100 + 100
			if end > len(bvids) {
				end = len(bvids)
			}
			bvidsSlice := bvids[start:end]
			rangeVideosInfo := []GroupVideoInfo{}
			db.Model(&models.VideosInfo{}).Select(
				"bvid", "title",
			).Where(
				"bvid IN (?)", bvidsSlice,
			).Group("bvid").Group("title").Find(&rangeVideosInfo)
			if len(rangeVideosInfo) > 0 {
				videoInfoList = append(videoInfoList, rangeVideosInfo...)
			}
		}
	}

	videoInfoMap := make(map[string]string)
	for _, v := range videoInfoList {
		videoInfoMap[v.Bvid] = v.Title
	}

	path := ""
	basePath := config.GetConfig().Download.Path

	if vType == consts.VIDEO_TYPE_COLLECTED {
		collect := models.CollectedInfo{}
		db.Model(&models.CollectedInfo{}).Where(&models.CollectedInfo{
			Mid: mid, CollId: source_id,
		}).First(&collect)
		if collect.ID > 0 {
			path = filepath.Join(utils.GetCollectedPath(mid, basePath), collect.Title)
		}
	} else if vType == consts.VIDEO_TYPE_WATCH_LATER {
		path = utils.GetWatchLaterPath(mid, basePath)
	} else if vType == consts.VIDEO_TYPE_FAVOUR {
		folder := models.FavourFoldersInfo{}
		db.Model(&models.FavourFoldersInfo{}).Where(&models.FavourFoldersInfo{
			Mid: mid, Mlid: source_id,
		}).First(&folder)
		if folder.ID > 0 {
			path = filepath.Join(utils.GetFavourPath(mid, basePath), folder.Title)
		}
	}
	if path == "" {
		return
	}
	for _, v := range videosList {
		if videoInfoTitle, ok := videoInfoMap[v.Bvid]; ok {
			beforePath := filepath.Join(path, utils.Name(videoInfoTitle))
			distPath := filepath.Join(path, utils.Name(videoInfoTitle)+"[已失效]")
			os.Rename(beforePath, distPath)
			logger.Infof("%s -> %s", beforePath, distPath)
		}
	}
}

func SetVideosInfo(videosInfo map[string]*VideoInfo) {
	db := models.GetDB()

	bvids := make([]string, 0)
	for _, v := range videosInfo {
		bvids = append(bvids, v.Bvid)
	}
	existsVideosInfo := []models.VideosInfo{}
	if len(bvids) > 0 {
		times := int(math.Ceil(float64(len(bvids)) / 100))
		for i := 0; i < times; i++ {
			start := i * 100
			end := i*100 + 100
			if end > len(bvids) {
				end = len(bvids)
			}
			bvidsSlice := bvids[start:end]
			rangeVideosInfo := []models.VideosInfo{}
			db.Model(&models.VideosInfo{}).Where(
				"bvid IN (?)", bvidsSlice,
			).Find(&rangeVideosInfo)
			if len(rangeVideosInfo) > 0 {
				existsVideosInfo = append(existsVideosInfo, rangeVideosInfo...)
			}
		}
	}

	existsVideosInfoMap := make(map[string]models.VideosInfo)
	for _, v := range existsVideosInfo {
		existsVideosInfoMap[fmt.Sprintf("%s_%d", v.Bvid, v.Cid)] = v
	}

	existsKeys := maps.Keys(existsVideosInfoMap)
	newKeys := maps.Keys(videosInfo)

	insertKeys := utils.Difference(newKeys, existsKeys)
	updateKeys := utils.Intersection(newKeys, existsKeys)

	if len(insertKeys) > 0 {
		createList := []*models.VideosInfo{}
		for _, key := range insertKeys {
			if info, ok := videosInfo[key]; ok {
				createList = append(createList, &models.VideosInfo{
					Bvid:   info.Bvid,
					Cid:    info.Cid,
					Title:  info.Title,
					Width:  info.Width,
					Height: info.Height,
					Rotate: info.Rotate,
					Page:   info.Page,
					Part:   info.Part,
				})
			}
		}
		if len(createList) > 0 {
			db.CreateInBatches(createList, 100)
		}
	}

	// TODO:信息先不删除，后面再想办法识别删除
	// deleteKeys := utils.Difference(existsKeys, newKeys)
	// if len(deleteKeys) > 0 {
	// 	for _, key := range deleteKeys {
	// 		if info, ok := existsVideosInfoMap[key]; ok {
	// 			db.Delete(&info)
	// 		}
	// 	}
	// }

	if len(updateKeys) > 0 {
		for _, key := range updateKeys {
			info, ok1 := videosInfo[key]
			if !ok1 {
				continue
			}
			existsInfo, ok := existsVideosInfoMap[key]
			if !ok {
				continue
			}
			if existsInfo.Title != info.Title || existsInfo.Width != info.Width || existsInfo.Height != info.Height || existsInfo.Rotate != info.Rotate || existsInfo.Page != info.Page || existsInfo.Part != info.Part {
				existsInfo.Part = info.Part
				existsInfo.Page = info.Page
				existsInfo.Title = info.Title
				existsInfo.Width = info.Width
				existsInfo.Height = info.Height
				existsInfo.Rotate = info.Rotate
				db.Save(existsInfo)
			}
		}
	}
}

func AfterRefresh(mid int) {
	db := models.GetDB()
	subQueryFav := db.Model(&models.FavourFoldersInfo{}).Where(
		&models.FavourFoldersInfo{Mid: mid, Sync: consts.FAVOUR_NEED_SYNC},
	).Select("mlid")
	subQueryCollected := db.Model(&models.CollectedInfo{}).Where(
		&models.CollectedInfo{Mid: mid, Sync: consts.COLLECTED_NEED_SYNC},
	).Select("coll_id")
	subQueryVideo := db.Model(&models.Videos{}).Where(
		"mid = ? AND status = ? AND ((source_id IN (?) AND type = ?) OR (source_id IN (?) AND type = ?))",
		mid, consts.VIDEO_STATUS_INIT, subQueryFav, consts.VIDEO_TYPE_FAVOUR, subQueryCollected, consts.VIDEO_TYPE_COLLECTED,
	).Select("id")

	db.Model(&models.Videos{}).Where(
		"id IN (?)", subQueryVideo,
	).Update("status", consts.VIDEO_STATUS_TO_BE_DOWNLOAD)

	var watchLaterSync int64
	db.Model(&models.WatchLater{}).Where(&models.WatchLater{Mid: mid, Sync: consts.WATCH_LATER_NEED_SYNC}).Count(&watchLaterSync)

	if watchLaterSync > 0 {
		db.Model(&models.Videos{}).Where(
			&models.Videos{
				Mid:    mid,
				Type:   consts.VIDEO_TYPE_WATCH_LATER,
				Status: consts.VIDEO_STATUS_INIT,
			},
		).Update("status", consts.VIDEO_STATUS_TO_BE_DOWNLOAD)
	}
}

func GetVideoInfo(bvid string, cid int) *VideoInfo {
	var videoInfo VideoInfo
	db := models.GetDB()
	db.Model(&models.VideosInfo{}).Where(&models.VideosInfo{Bvid: bvid, Cid: cid}).First(&videoInfo)
	return &videoInfo
}
