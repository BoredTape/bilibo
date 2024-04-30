package services

import (
	"bilibo/consts"
	"bilibo/models"
	"fmt"

	"golang.org/x/exp/maps"
)

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

	query := db.Model(&models.FavourVideos{}).Where("status IN (?)", statusList)

	query.Count(&total)

	if total > 0 {
		var favourVideos []models.FavourVideos
		query.Order("updated_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&favourVideos)
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
