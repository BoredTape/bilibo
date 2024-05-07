package services

import (
	"bilibo/consts"
	"bilibo/models"
	"fmt"
	"slices"

	"golang.org/x/exp/maps"
)

func DelVideoByMid(mid int) {
	db := models.GetDB()
	db.Where(models.Videos{Mid: mid}).Delete(&models.Videos{})
}

type VideoInfo struct {
	Part        string `json:"part"`
	Title       string `json:"title"`
	Bvid        string `json:"bvid"`
	Status      int    `json:"status"`
	SourceId    int    `json:"source_id"`
	SourceTitle string `json:"source_title"`
	Mid         int    `json:"mid"`
	AccountName string `json:"account_name"`
	Type        int    `json:"type"`
}

func handleQueryStatus(status int) []int {
	statusList := []int{status}
	if status == consts.VIDEO_STATUS_DOWNLOAD_FAIL {
		statusList = append(statusList, consts.VIDEO_STATUS_DOWNLOAD_RETRY)
	}
	return statusList

}

func GetVideosByStatus(status, page, pageSize int) (*[]*VideoInfo, int64) {
	result := make([]*VideoInfo, 0)
	db := models.GetDB()
	var total int64

	statusList := handleQueryStatus(status)

	query := db.Model(&models.Videos{}).Where("status IN (?)", statusList)

	query.Count(&total)

	if total > 0 {
		var videos []models.Videos
		query.Order("updated_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&videos)
		accountMap := make(map[int]*AccountInfo)
		favMap := make(map[int]*FavourFolders)
		collectedMap := make(map[int]*Collected)

		videoBvids := make([]string, 0)
		for _, v := range videos {
			accountMap[v.Mid] = nil
			if v.Type == consts.VIDEO_TYPE_FAVOUR {
				favMap[v.SourceId] = nil
			}
			if !slices.Contains(videoBvids, v.Bvid) {
				videoBvids = append(videoBvids, v.Bvid)
			}
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

		var collectedInfos []models.CollectedInfo
		db.Where("mid IN (?)", maps.Keys(accountMap)).Find(&collectedInfos)
		for _, v := range collectedInfos {
			collectedMap[v.Mid] = &Collected{
				CollId:     v.CollId,
				Attr:       v.Attr,
				Title:      v.Title,
				MediaCount: v.MediaCount,
				Sync:       v.Sync,
			}
		}

		var videosInfo []models.VideosInfo
		db.Model(&models.VideosInfo{}).Where("bvid IN (?)", videoBvids).Find(&videosInfo)
		videoMap := make(map[string]*models.VideosInfo)
		for _, v := range videosInfo {
			videoMap[fmt.Sprintf("%s_%d", v.Bvid, v.Cid)] = &v
		}

		for _, v := range videos {
			videoInfo := videoMap[fmt.Sprintf("%s_%d", v.Bvid, v.Cid)]
			sourceTitle := ""
			if v.Type == consts.VIDEO_TYPE_FAVOUR {
				if favTitle, ok := favMap[v.SourceId]; ok {
					sourceTitle = fmt.Sprintf("收藏夹：%s", favTitle.Title)
				}
			} else if v.Type == consts.VIDEO_TYPE_WATCH_LATER {
				sourceTitle = consts.ACCOUNT_DIR_WATCH_LATER
			} else if v.Type == consts.VIDEO_TYPE_COLLECTED {
				if collTitle, ok := collectedMap[v.Mid]; ok {
					sourceTitle = fmt.Sprintf("订阅：%s", collTitle.Title)
				}
			}

			accountName := ""
			if accountMap[v.Mid] != nil {
				accountName = accountMap[v.Mid].Uname
			}
			result = append(result, &VideoInfo{
				Part:        fmt.Sprintf("P%d %s", videoInfo.Page, videoInfo.Part),
				Title:       videoInfo.Title,
				Bvid:        v.Bvid,
				Status:      v.Status,
				SourceId:    v.SourceId,
				SourceTitle: sourceTitle,
				Mid:         v.Mid,
				AccountName: accountName,
				Type:        v.Type,
			})
		}
	}

	return &result, total
}

func SetToViewStatus(mid int, status int) {
	db := models.GetDB()
	db.Model(&models.Videos{}).Where(
		"mid = ? AND type = ?", consts.VIDEO_TYPE_WATCH_LATER,
	).Update("status", status)
}
