package services

import (
	"bilibo/consts"
	"bilibo/models"
)

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

func DelQRCodeInfo(qrId string) {
	db := models.GetDB()
	db.Where(models.QRCode{QRID: qrId}).Delete(&models.QRCode{})
}

func ClearAllQRCode() {
	db := models.GetDB()
	db.Where("deleted_at IS NULL").Delete(&models.QRCode{})
}

func GetAccountList() *[]models.BiliAccounts {
	db := models.GetDB()
	var account []models.BiliAccounts
	db.Model(&models.BiliAccounts{
		Status: consts.ACCOUNT_STATUS_NORMAL,
	}).Find(&account)
	return &account
}
