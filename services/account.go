package services

import (
	"bilibo/consts"
	"bilibo/models"
	"fmt"
	"net/url"

	"golang.org/x/exp/maps"
)

func SaveAccountInfo(mid int, uname, face, cookies, imgKey, subKey string) {
	db := models.GetDB()
	account := models.BiliAccounts{}
	db.Where(models.BiliAccounts{Mid: mid}).FirstOrInit(&account)
	account.Cookies = cookies
	account.ImgKey = imgKey
	account.SubKey = subKey
	account.Status = consts.ACCOUNT_STATUS_NORMAL
	account.UName = uname
	account.Face = face
	db.Save(&account)
}

func UpdateAccountWBI(mid int, imgKey, subKey string) {
	db := models.GetDB()
	account := models.BiliAccounts{}
	db.Where(models.BiliAccounts{Mid: mid}).First(&account)
	if account.ID == 0 {
		return
	}
	if account.ImgKey == imgKey && account.SubKey == subKey {
		return
	}
	account.ImgKey = imgKey
	account.SubKey = subKey
	db.Save(&account)
}

func SetAccountStatus(mid int, status int) {
	db := models.GetDB()
	var account models.BiliAccounts
	db.Where(models.BiliAccounts{Mid: mid}).First(&account)
	account.Status = status
	if account.ID > 0 {
		db.Save(&account)
	}
}

func DelAccount(mid int) {
	db := models.GetDB()
	db.Where(models.BiliAccounts{Mid: mid}).Delete(&models.BiliAccounts{})
}

func AddQRCodeInfo(qrId string) {
	db := models.GetDB()
	qrcode := models.QRCode{QRID: qrId}
	qrcode.Status = consts.QRCODE_STATUS_NOT_SCAN
	db.Save(&qrcode)
}

func GetQRCodeInfo(qrId string) *models.QRCode {
	db := models.GetDB()
	var qrcode models.QRCode
	db.Where(models.QRCode{QRID: qrId}).First(&qrcode)
	return &qrcode
}

func DelQRCodeInfo(qrId string) {
	db := models.GetDB()
	db.Where(models.QRCode{QRID: qrId}).Delete(&models.QRCode{})
}

func SetQRCodeStatus(qrId string, status int) {
	db := models.GetDB()
	var qrcode models.QRCode
	db.Where(models.QRCode{QRID: qrId}).First(&qrcode)
	qrcode.Status = status
	db.Save(&qrcode)
}

func ClearAllQRCode() {
	db := models.GetDB()
	db.Where("deleted_at IS NULL").Delete(&models.QRCode{})
}

type AccountInfo struct {
	Mid          int              `json:"mid"`
	Uname        string           `json:"uname"`
	Status       int              `json:"status"`
	Face         string           `json:"face"`
	FoldersCount int              `json:"folders_count"`
	Folders      []*FavourFolders `json:"folders"`
}

func AccountList(page, pageSize int) (*[]*AccountInfo, int64) {
	db := models.GetDB()
	accountMap := make(map[int]*AccountInfo, 0)
	total := AccountTotal()
	if total > 0 {
		var datas []models.BiliAccounts
		db.Model(&models.BiliAccounts{}).Limit(pageSize).Offset((page - 1) * pageSize).Find(&datas)
		for _, data := range datas {
			item := AccountInfo{
				Mid:    data.Mid,
				Status: data.Status,
				Face: fmt.Sprintf(
					"/api/account/proxy/%d/?url=%s",
					data.Mid,
					url.QueryEscape(data.Face),
				),
				Uname:        data.UName,
				Folders:      make([]*FavourFolders, 0),
				FoldersCount: 0,
			}
			accountMap[data.Mid] = &item
		}

		var favourFolderInfos []models.FavourFoldersInfo
		db.Where("mid IN (?)", maps.Keys(accountMap)).Find(&favourFolderInfos)
		for _, v := range favourFolderInfos {
			folders := FavourFolders{
				Mlid:       v.Mlid,
				Fid:        v.Fid,
				Title:      v.Title,
				MediaCount: v.MediaCount,
				Sync:       v.Sync,
			}
			accountMap[v.Mid].Folders = append(accountMap[v.Mid].Folders, &folders)
			accountMap[v.Mid].FoldersCount++
		}
	}
	items := maps.Values(accountMap)
	return &items, total
}

func AccountTotal() int64 {
	db := models.GetDB()
	var total int64
	db.Model(&models.BiliAccounts{}).Count(&total)
	return total
}
