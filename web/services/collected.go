package services

import (
	"bilibo/consts"
	"bilibo/models"
)

func DelCollectedByMid(mid int) {
	db := models.GetDB()
	db.Where("mid = ?", mid).Delete(&models.CollectedInfo{})
}

func SetCollectedSyncStatus(mid, collId, status int) {
	db := models.GetDB()
	db.Model(&models.CollectedInfo{}).Where(
		&models.CollectedInfo{CollId: collId, Mid: mid},
	).Update("sync", status)
	if status == consts.COLLECTED_NEED_SYNC {
		db.Model(&models.Videos{}).Where(
			"mid = ? AND source_id = ? AND status = ? AND type = ?",
			mid, collId, consts.VIDEO_STATUS_INIT, consts.VIDEO_TYPE_COLLECTED,
		).Update("status", consts.VIDEO_STATUS_TO_BE_DOWNLOAD)
	} else if status == consts.COLLECTED_NOT_SYNC {
		db.Model(&models.Videos{}).Where(
			"mid = ? AND source_id = ? AND status = ? AND type = ?",
			mid, collId, consts.VIDEO_STATUS_TO_BE_DOWNLOAD, consts.VIDEO_TYPE_COLLECTED,
		).Update("status", consts.VIDEO_STATUS_INIT)
	}
}

func GetAccountCollectIdInfoByMid(mid int) *[]*Collected {
	db := models.GetDB()
	var infos []models.CollectedInfo
	db.Model(&models.CollectedInfo{}).Where("mid = ?", mid).Order("coll_id DESC").Find(&infos)
	datas := make([]*Collected, 0)
	for _, v := range infos {
		datas = append(datas, &Collected{
			CollId:     v.CollId,
			Title:      v.Title,
			Attr:       v.Attr,
			MediaCount: v.MediaCount,
			Sync:       v.Sync,
		})
	}
	return &datas
}
