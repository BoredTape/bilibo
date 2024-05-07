package services

import (
	"bilibo/config"
	"bilibo/consts"
	"bilibo/models"
	"bilibo/universal"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/maruel/natural"
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

func DelAccount(mid int) {
	db := models.GetDB()
	account := models.BiliAccounts{}
	db.Model(&models.BiliAccounts{}).Where("mid = ?", mid).Find(&account)
	if account.ID < 1 {
		return
	}
	*universal.GetCH() <- universal.CH{
		Mid:     account.Mid,
		UName:   account.UName,
		Face:    account.Face,
		ImgKey:  account.ImgKey,
		SubKey:  account.SubKey,
		Cookies: account.Cookies,
		Action:  consts.CHANNEL_ACTION_ADD_CLIENT,
	}
	db.Delete(&account)
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

func SetQRCodeStatus(qrId string, status int) {
	db := models.GetDB()
	var qrcode models.QRCode
	db.Where(models.QRCode{QRID: qrId}).First(&qrcode)
	qrcode.Status = status
	db.Save(&qrcode)
}

type FavourFolders struct {
	Mlid       int    `json:"mlid"`
	Fid        int    `json:"fid"`
	Title      string `json:"title"`
	MediaCount int    `json:"media_count"`
	Sync       int    `json:"sync"`
}

type Collected struct {
	CollId     int    `json:"coll_id"`
	Attr       int    `json:"attr"`
	Title      string `json:"title"`
	MediaCount int    `json:"media_count"`
	Sync       int    `json:"sync"`
}

type AccountInfo struct {
	Mid             int              `json:"mid"`
	Uname           string           `json:"uname"`
	Status          int              `json:"status"`
	Face            string           `json:"face"`
	FoldersCount    int              `json:"folders_count"`
	Folders         []*FavourFolders `json:"folders"`
	WatchLaterCount int64            `json:"watch_later_count"`
	WatchLaterSync  int              `json:"watch_later_sync"`
	Collected       []*Collected     `json:"collected"`
	CollectedCount  int              `json:"collected_count"`
}

type AccountWatchLaterCount struct {
	Mid   int   `json:"mid"`
	Count int64 `json:"count"`
}

type AccountWatchLaterSync struct {
	Mid  int `json:"mid"`
	Sync int `json:"sync"`
}

func AccountList(page, pageSize int) (*[]*AccountInfo, int64) {
	db := models.GetDB()
	accountMap := make(map[int]*AccountInfo)
	accountMids := make([]int, 0)
	total := AccountTotal()
	if total > 0 {
		var datas []models.BiliAccounts
		db.Model(&models.BiliAccounts{}).Order("updated_at DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&datas)
		for _, data := range datas {
			item := AccountInfo{
				Mid:    data.Mid,
				Status: data.Status,
				Face: fmt.Sprintf(
					"/api/account/proxy/%d/?url=%s",
					data.Mid,
					url.QueryEscape(data.Face),
				),
				Uname:           data.UName,
				Folders:         make([]*FavourFolders, 0),
				FoldersCount:    0,
				WatchLaterCount: 0,
				WatchLaterSync:  0,
			}
			accountMap[data.Mid] = &item
			accountMids = append(accountMids, data.Mid)
		}

		var favourFolderInfos []models.FavourFoldersInfo
		db.Model(&models.FavourFoldersInfo{}).Where("mid IN (?)", accountMids).Find(&favourFolderInfos)
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

		var collectedInfos []models.CollectedInfo
		db.Model(&models.CollectedInfo{}).Where("mid IN (?)", accountMids).Find(&collectedInfos)
		for _, v := range collectedInfos {
			collected := Collected{
				CollId:     v.CollId,
				Attr:       v.Attr,
				Title:      v.Title,
				MediaCount: v.MediaCount,
				Sync:       v.Sync,
			}
			accountMap[v.Mid].Collected = append(accountMap[v.Mid].Collected, &collected)
			accountMap[v.Mid].CollectedCount++
		}

		watchLaterCount := make([]AccountWatchLaterCount, 0)
		db.Model(&models.Videos{}).Select("COUNT(mid) AS count", "mid").Where(
			"mid IN (?) AND type = ?", accountMids,
			consts.VIDEO_TYPE_WATCH_LATER,
		).Group("mid").Find(&watchLaterCount)
		for _, v := range watchLaterCount {
			accountMap[v.Mid].WatchLaterCount = v.Count
		}

		watchLaterSync := make([]AccountWatchLaterSync, 0)
		db.Model(&models.WatchLater{}).Select("sync", "mid").Where(
			"mid IN (?)", accountMids,
		).Find(&watchLaterSync)
		for _, v := range watchLaterSync {
			accountMap[v.Mid].WatchLaterSync = v.Sync
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

type AccountFile struct {
	BaseName      string   `json:"basename"`
	Extension     string   `json:"extension"`
	ExtraMetadata []string `json:"extra_metadata"`
	FileSize      int64    `json:"file_size"`
	LastModified  int64    `json:"last_modified"`
	MimeType      *string  `json:"mime_type"`
	Path          string   `json:"path"`
	Storage       string   `json:"storage"`
	Type          string   `json:"type"`
	Visibility    string   `json:"visibility"`
}

func GetAccountIndex(mid, action, path string) map[string]interface{} {
	result := make(map[string]interface{})
	conf := config.GetConfig()
	rootPath := filepath.Join(conf.Download.Path, mid)

	result["adapter"] = mid
	result["dirname"] = path
	result["storages"] = []string{mid}

	files := make([]*AccountFile, 0)
	result["files"] = files

	subPath := filepath.Join(rootPath, strings.ReplaceAll(path, mid+"://", "/"))
	fileMap := make(map[string]*AccountFile)
	fileNames := make([]string, 0)
	dirFiles, err := os.ReadDir(subPath)
	if err != nil {
		return result
	}

	for _, file := range dirFiles {
		if fileInfo, err := file.Info(); err == nil {
			file := AccountFile{
				Path:          mid + ":/" + filepath.Join(strings.ReplaceAll(path, mid+"://", "/"), fileInfo.Name()),
				Visibility:    "public",
				ExtraMetadata: make([]string, 0),
				FileSize:      fileInfo.Size(),
				LastModified:  fileInfo.ModTime().Unix(),
				Storage:       mid,
				BaseName:      fileInfo.Name(),
				MimeType:      nil,
			}
			if fileInfo.IsDir() {
				file.Type = "dir"
				file.Extension = ""
			} else {
				file.Type = "file"
			}
			fileNames = append(fileNames, fileInfo.Name())
			fileMap[fileInfo.Name()] = &file
		}
	}

	if len(fileNames) < 1 {
		return result
	}

	sort.Sort(natural.StringSlice(fileNames))

	for _, fileName := range fileNames {
		file := fileMap[fileName]
		if file.Type == "file" {
			mtype, err := mimetype.DetectFile(filepath.Join(subPath, file.BaseName))
			if err != nil {
				file.MimeType = nil
				fextension := strings.Split(file.BaseName, ".")
				slices.Reverse(fextension)
				file.Extension = fextension[0]
				continue
			} else {
				fmtype := mtype.String()
				file.MimeType = &fmtype
				file.Extension = strings.Replace(mtype.Extension(), ".", "", 1)
			}
		}
		files = append(files, file)
	}

	result["files"] = files
	return result
}
func GetAccountFileDownload(mid, action, path string) (string, error) {
	conf := config.GetConfig()
	rootPath := filepath.Join(conf.Download.Path, mid)
	filePath := filepath.Join(rootPath, strings.ReplaceAll(path, mid+"://", "/"))
	if _, err := os.Stat(filePath); err != nil {
		return "", err
	} else {
		return filePath, nil
	}
}
