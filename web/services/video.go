package services

import (
	"bilibo/consts"
	"bilibo/models"
	"fmt"

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
		accountMap := make(map[int]*AccountInfo, 0)
		favMap := make(map[int]*FavourFolders, 0)
		for _, v := range videos {
			accountMap[v.Mid] = nil
			if v.Type == consts.VIDEO_TYPE_FAVOUR {
				favMap[v.SourceId] = nil
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

		for _, v := range videos {
			sourceTitle := ""
			if v.Type == consts.VIDEO_TYPE_FAVOUR {
				if favMap[v.SourceId] != nil {
					sourceTitle = fmt.Sprintf("收藏夹：%s", favMap[v.SourceId].Title)
				}
			} else if v.Type == consts.VIDEO_TYPE_WATCH_LATER {
				sourceTitle = consts.ACCOUNT_DIR_WATCH_LATER
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
