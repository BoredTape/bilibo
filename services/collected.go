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

type Collected struct {
	Id         int
	Mid        int
	Attr       int
	Title      string
	MediaCount int
}
type CollectedInfo struct {
	Count int
	List  []Collected
}

func SetCollectedInfo(mid int, collectedInfo *CollectedInfo) {
	if collectedInfo == nil {
		return
	}
	db := models.GetDB()
	var existCollectedInfos []models.CollectedInfo
	db.Model(&models.CollectedInfo{}).Where("mid = ?", mid).Find(&existCollectedInfos)
	existMap := make(map[int]models.CollectedInfo)
	for _, v := range existCollectedInfos {
		existMap[v.CollId] = v
	}
	existCollIds := maps.Keys(existMap)
	accountCollIds := make([]int, 0)

	insertList := make([]*models.CollectedInfo, 0)
	updateList := make([]*models.CollectedInfo, 0)
	deleteCollIds := make([]int, 0)

	for _, v := range collectedInfo.List {
		accountCollIds = append(accountCollIds, v.Id)
		if !slices.Contains(existCollIds, v.Id) {
			insertList = append(insertList, &models.CollectedInfo{
				CollId:     v.Id,
				Mid:        mid,
				Attr:       v.Attr,
				Title:      v.Title,
				MediaCount: v.MediaCount,
				Sync:       consts.COLLECTED_NOT_SYNC,
			})
		} else if slices.Contains(existCollIds, v.Id) {
			existInfo := existMap[v.Id]
			if existInfo.Attr != v.Attr || existInfo.Title != v.Title || existInfo.MediaCount != v.MediaCount {
				updateList = append(updateList, &models.CollectedInfo{
					CollId:     v.Id,
					Mid:        mid,
					Attr:       v.Attr,
					Title:      v.Title,
					MediaCount: v.MediaCount,
					Sync:       existInfo.Sync,
				})
			}
		}
	}

	for _, v := range existCollIds {
		if !slices.Contains(accountCollIds, v) {
			deleteCollIds = append(deleteCollIds, v)
		}
	}

	if len(insertList) > 0 {
		for _, insert_data := range insertList {
			db.Model(&models.CollectedInfo{}).Create(insert_data)
		}
	}

	if len(deleteCollIds) > 0 {
		go DeleteCollecteds(mid, deleteCollIds)
	}

	if len(updateList) > 0 {
		conf := config.GetConfig()
		for _, updateData := range updateList {
			existInfo := existMap[updateData.CollId]
			oldTitle := utils.Name(existInfo.Title)
			newTitle := utils.Name(updateData.Title)
			if newTitle != oldTitle {
				favPath := utils.GetCollectedPath(existInfo.Mid, conf.Download.Path)
				oldPath := filepath.Join(favPath, oldTitle)
				newPath := filepath.Join(favPath, newTitle)
				go ChangeCollectedName(updateData.CollId, oldPath, newPath)
			}
			db.Model(&models.CollectedInfo{}).Where("id = ?", existInfo.ID).Updates(updateData)
		}
	}
}

func GetCollectedInfoByCollidMid(mid, collId int) *models.CollectedInfo {
	db := models.GetDB()
	var collectedInfo models.CollectedInfo
	db.Where("coll_id = ? AND mid = ?", collId, mid).First(&collectedInfo)
	return &collectedInfo
}

func ChangeCollectedName(collId int, oldPath, newPath string) {
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
	db.Exec(sqlPause, value, collId, consts.VIDEO_TYPE_COLLECTED)
	t := NewTask(
		WithTaskType(consts.TASK_TYPE_RUNNING_TIME),
		WithName("更改收藏和订阅名字: "+oldPath+" => "+newPath),
		WithTaskId(fmt.Sprintf("change_collected_name_%d", collId)),
	)
	t.Save()
	for {
		logger.Info(fmt.Sprintf("更改收藏和订阅路径更改:\n%s => %s", oldPath, newPath))
		var downloadingCount int64
		db.Model(&models.Videos{}).Where(
			"source_id = ? AND status = ? AND type = ?",
			collId, consts.VIDEO_STATUS_DOWNLOADING, consts.VIDEO_TYPE_COLLECTED,
		).Count(&downloadingCount)
		fmt.Println(downloadingCount)
		if downloadingCount == 0 {
			fmt.Println(oldPath, "\n", newPath)
			if err := utils.RenameDir(oldPath, newPath); err != nil {
				logger.Error(err.Error())
			}
			sqlContinue := "UPDATE " + table.TableName() + " SET status=status-100 WHERE status > 100 AND deleted_at IS NULL AND source_id = ? AND type = ?;"
			db.Exec(sqlContinue, collId, consts.VIDEO_TYPE_COLLECTED)
			break
		} else {
			logger.Info(fmt.Sprintf("收藏和订阅路径 %s 正在下载,重试中...", oldPath))
		}
		t.UpdateNextRunningAt(2)
		time.Sleep(2 * time.Second)
	}
	t.Delete()
}

func DeleteCollecteds(mid int, collIds []int) {
	db := models.GetDB()
	logger := log.GetLogger()
	collectedInfos := []models.CollectedInfo{}
	db.Where("coll_id IN (?) AND mid = ?", collIds, mid).Find(&collectedInfos)
	conf := config.GetConfig()
	basePath := conf.Download.Path
	db.Where(
		"source_id IN (?) AND type = ? AND status != ? AND mid = ?",
		collIds, consts.VIDEO_TYPE_COLLECTED,
		consts.VIDEO_STATUS_DOWNLOADING, mid,
	).Delete(&models.Videos{})

	collIdsStr := strings.Trim(strings.Replace(fmt.Sprint(collIds), " ", ",", -1), "[]")
	fullCollIdsStr := fmt.Sprintf("删除收藏和订阅,收藏和订阅IDs:[%s]", collIdsStr)
	t := NewTask(
		WithTaskType(consts.TASK_TYPE_RUNNING_TIME),
		WithName(fullCollIdsStr),
		WithTaskId(fmt.Sprintf("delete_collected:%s", collIdsStr)),
	)
	t.Save()
	for {
		var downloadingCount int64
		db.Model(&models.Videos{}).Where(
			"source_id IN (?) AND type = ? AND status = ? AND mid = ?",
			collIds, consts.VIDEO_TYPE_COLLECTED,
			consts.VIDEO_STATUS_DOWNLOADING, mid,
		).Count(&downloadingCount)
		if downloadingCount == 0 {
			db.Where("coll_id IN (?) AND mid = ?", collIds, mid).Delete(&models.CollectedInfo{})
			db.Where(
				"source_id IN (?) AND type = ? AND mid = ?",
				collIds, consts.VIDEO_TYPE_COLLECTED, mid,
			).Delete(&models.Videos{})
			for _, collected := range collectedInfos {
				utils.RecyclePath(collected.Mid, basePath, utils.GetCollectedPath(collected.Mid, basePath), utils.Name(collected.Title))
			}
			break
		} else {
			logger.Info("收藏和订阅视频正在下载,重试中...")
		}
		t.UpdateNextRunningAt(2)
		time.Sleep(2 * time.Second)
	}
	t.Delete()
}
