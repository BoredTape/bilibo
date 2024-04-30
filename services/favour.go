package services

import (
	"bilibo/bili/bili_client"
	"bilibo/config"
	"bilibo/consts"
	"bilibo/models"
	"bilibo/utils"
	"path/filepath"
	"slices"
	"strings"

	"bilibo/realtime_job"

	"golang.org/x/exp/maps"
)

func SetFavourInfo(mid int, favInfo *bili_client.AllFavourFolderInfo) {
	if favInfo == nil {
		return
	}
	db := models.GetDB()
	var existFavourInfos []models.FavourFoldersInfo
	db.Model(&models.FavourFoldersInfo{}).Where("mid = ?", mid).Find(&existFavourInfos)
	existMap := make(map[int]models.FavourFoldersInfo)
	for _, v := range existFavourInfos {
		existMap[v.Mlid] = v
	}
	existMlids := maps.Keys(existMap)

	insertList := make([]*models.FavourFoldersInfo, 0)
	updateList := make([]*models.FavourFoldersInfo, 0)
	deleteMlids := make([]int, 0)

	for _, v := range favInfo.List {
		if !slices.Contains(existMlids, v.Id) {
			insertList = append(insertList, &models.FavourFoldersInfo{
				Mid:        mid,
				Fid:        v.Fid,
				MediaCount: v.MediaCount,
				Attr:       v.Attr,
				Title:      v.Title,
				Mlid:       v.Id,
				FavState:   v.FavState,
				Sync:       consts.FAVOUR_NOT_SYNC,
			})
		} else if slices.Contains(existMlids, v.Id) {
			existInfo := existMap[v.Id]
			if existInfo.Attr != v.Attr || existInfo.Title != v.Title || existInfo.FavState != v.FavState || existInfo.MediaCount != v.MediaCount {
				updateList = append(updateList, &models.FavourFoldersInfo{
					MediaCount: v.MediaCount,
					Attr:       v.Attr,
					Title:      v.Title,
					FavState:   v.FavState,
				})
			}
		} else {
			deleteMlids = append(deleteMlids, v.Id)
		}
	}

	if len(insertList) > 0 {
		db.Create(insertList)
	}

	if len(deleteMlids) > 0 {
		go realtime_job.DeleteFavours(deleteMlids)
	}

	if len(updateList) > 0 {
		conf := config.GetConfig()
		for _, updateData := range updateList {
			existInfo := existMap[updateData.Mlid]
			oldTitle := strings.ReplaceAll(existInfo.Title, "/", "⁄")
			newTitle := strings.ReplaceAll(updateData.Title, "/", "⁄")
			if newTitle != oldTitle {
				favPath := utils.GetFavourPath(existInfo.Mid, conf.Download.Path)
				oldPath := filepath.Join(favPath, oldTitle)
				newPath := filepath.Join(favPath, newTitle)
				go realtime_job.ChangeFavourName(updateData.Mlid, oldPath, newPath)
			}
			db.Model(&models.FavourFoldersInfo{}).Where("id = ?", existInfo.ID).Updates(updateData)
		}
	}
}

func GetFavourInfoByMlid(mlid int) *models.FavourFoldersInfo {
	db := models.GetDB()
	var favourFolderInfo models.FavourFoldersInfo
	db.Where(models.FavourFoldersInfo{Mlid: mlid}).First(&favourFolderInfo)
	if favourFolderInfo.ID == 0 {
		return nil
	}
	return &favourFolderInfo
}
