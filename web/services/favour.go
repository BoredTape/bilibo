package services

import (
	"bilibo/consts"
	"bilibo/models"
)

func DelFavourInfoByMid(mid int) {
	db := models.GetDB()
	db.Where(models.FavourFoldersInfo{Mid: mid}).Delete(&models.FavourFoldersInfo{})
}

func GetAccountFavourInfoByMid(mid int) *[]*FavourFolders {
	db := models.GetDB()
	var favourFolderInfos []models.FavourFoldersInfo
	db.Model(&models.FavourFoldersInfo{}).Where("mid = ?", mid).Order("mlid DESC").Find(&favourFolderInfos)
	datas := make([]*FavourFolders, 0)
	for _, v := range favourFolderInfos {
		datas = append(datas, &FavourFolders{
			Mlid:       v.Mlid,
			Fid:        v.Fid,
			Title:      v.Title,
			MediaCount: v.MediaCount,
			Sync:       v.Sync,
		})
	}
	return &datas
}
func SetFavourSyncStatus(mid, mlid, status int) {
	db := models.GetDB()
	db.Model(&models.FavourFoldersInfo{}).Where(
		&models.FavourFoldersInfo{Mlid: mlid, Mid: mid},
	).Update("sync", status)
	if status == consts.FAVOUR_NEED_SYNC {
		db.Model(&models.Videos{}).Where(
			"mid = ? AND source_id = ? AND status = ? AND type = ?",
			mid, mlid, consts.VIDEO_STATUS_INIT, consts.VIDEO_TYPE_FAVOUR,
		).Update("status", consts.VIDEO_STATUS_TO_BE_DOWNLOAD)
	} else if status == consts.FAVOUR_NOT_SYNC {
		db.Model(&models.Videos{}).Where(
			"mid = ? AND source_id = ? AND status = ? AND type = ?",
			mid, mlid, consts.VIDEO_STATUS_TO_BE_DOWNLOAD, consts.VIDEO_TYPE_FAVOUR,
		).Update("status", consts.VIDEO_STATUS_INIT)
	}
}
