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
	db.Model(&models.FavourFoldersInfo{}).Where("mlid = ?", mlid).Update("sync", status)
	if status == consts.FAVOUR_NEED_SYNC {
		db.Model(&models.Videos{}).Where("mid = ? AND mlid = ? AND status = ?", mid, mlid, consts.VIDEO_STATUS_INIT).Update("status", consts.VIDEO_STATUS_TO_BE_DOWNLOAD)
	}
}
