package services

import (
	"bilibo/config"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/models"
	"bilibo/utils"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

type FolderInfo struct {
	Id         int
	Fid        int
	Mid        int
	Attr       int
	Title      string
	FavState   int
	MediaCount int
}
type FavourFolderInfo struct {
	Count int
	List  []FolderInfo
}

func SetFavourInfo(mid int, favInfo *FavourFolderInfo) {
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
	accountMlids := make([]int, 0)

	insertList := make([]*models.FavourFoldersInfo, 0)
	updateList := make([]*models.FavourFoldersInfo, 0)
	deleteMlids := make([]int, 0)

	for _, v := range favInfo.List {
		accountMlids = append(accountMlids, v.Id)
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
					Mlid:       v.Id,
					MediaCount: v.MediaCount,
					Attr:       v.Attr,
					Title:      v.Title,
					FavState:   v.FavState,
				})
			}
		}
	}

	for _, v := range existMlids {
		if !slices.Contains(accountMlids, v) {
			deleteMlids = append(deleteMlids, v)
		}
	}

	if len(insertList) > 0 {
		for _, insert_data := range insertList {
			db.Model(&models.FavourFoldersInfo{}).Create(insert_data)
		}
	}

	if len(deleteMlids) > 0 {
		go DeleteFavours(deleteMlids)
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
				go ChangeFavourName(updateData.Mlid, oldPath, newPath)
			}
			db.Model(&models.FavourFoldersInfo{}).Where("id = ?", existInfo.ID).Updates(updateData)
		}
	}
}

func ChangeFavourName(mlid int, oldPath, newPath string) {
	db := models.GetDB()
	logger := log.GetLogger()
	table := models.Videos{}
	sqlPause := "UPDATE " + table.TableName() +
		" SET status=status+100 WHERE status IN (?) AND source_id = ? AND type = ? AND deleted_at IS NULL;"
	value := []int{
		consts.VIDEO_STATUS_TO_BE_DOWNLOAD,
		consts.VIDEO_STATUS_DOWNLOAD_FAIL,
		consts.VIDEO_STATUS_DOWNLOAD_RETRY,
	}
	db.Exec(sqlPause, value, mlid, consts.VIDEO_TYPE_FAVOUR)
	for {
		t := NewTask(
			WithTaskType(consts.TASK_TYPE_RUNNING_TIME),
			WithName("更改收藏夹名字: "+oldPath+" => "+newPath),
			WithTaskId(fmt.Sprintf("change_favour_name_%d", mlid)),
		)
		t.Save()
		logger.Info(fmt.Sprintf("收藏夹路径更改:\n%s => %s", oldPath, newPath))
		var downloadingCount int64
		db.Model(&models.Videos{}).Where(
			"source_id = ? AND status = ? AND type = ?",
			mlid, consts.VIDEO_STATUS_DOWNLOADING, consts.VIDEO_TYPE_FAVOUR,
		).Count(&downloadingCount)
		fmt.Println(downloadingCount)
		if downloadingCount == 0 {
			fmt.Println(oldPath, "\n", newPath)
			if err := utils.RenameDir(oldPath, newPath); err != nil {
				logger.Error(err.Error())
			}
			sqlContinue := "UPDATE " + table.TableName() + " SET status=status-100 WHERE status > 100 AND deleted_at IS NULL AND source_id = ? AND type = ?;"
			db.Exec(sqlContinue, mlid, consts.VIDEO_TYPE_FAVOUR)
			break
		} else {
			logger.Info(fmt.Sprintf("收藏夹路径 %s 正在下载,重试中...", oldPath))
		}
		t.UpdateNextRunningAt(2)
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
		"source_id IN (?) AND type = ? AND status != ?",
		mlids, consts.VIDEO_TYPE_FAVOUR,
		consts.VIDEO_STATUS_DOWNLOADING,
	).Delete(&models.Videos{})
	for {
		mlidsStr := strings.Trim(strings.Replace(fmt.Sprint(mlids), " ", ",", -1), "[]")
		fullMlidsStr := fmt.Sprintf("删除收藏夹,收藏夹IDs:[%s]", mlidsStr)
		t := NewTask(
			WithTaskType(consts.TASK_TYPE_RUNNING_TIME),
			WithName(fullMlidsStr),
			WithTaskId(fmt.Sprintf("delete_favours:%s", mlidsStr)),
		)
		t.Save()
		logger.Info(fullMlidsStr)
		var downloadingCount int64
		db.Model(&models.Videos{}).Where(
			"source_id IN (?) AND type = ? AND status = ?",
			mlids, consts.VIDEO_TYPE_FAVOUR,
			consts.VIDEO_STATUS_DOWNLOADING,
		).Count(&downloadingCount)
		if downloadingCount == 0 {
			db.Where("mlid IN (?)", mlids).Delete(&models.FavourFoldersInfo{})
			db.Where(
				"source_id IN (?) AND type = ?",
				mlids, consts.VIDEO_TYPE_FAVOUR,
			).Delete(&models.Videos{})
			for _, fav := range favInfos {
				utils.RecyclePath(fav.Mid, basePath, utils.Name(fav.Title))
			}
			break
		} else {
			logger.Info("收藏夹视频正在下载,重试中...")
		}
		t.UpdateNextRunningAt(2)
		time.Sleep(2 * time.Second)
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
