package services

import (
	"bilibo/bili/bili_client"
	"bilibo/config"
	"bilibo/consts"
	"bilibo/models"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/maruel/natural"
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
		db.Model(&models.FavourFoldersInfo{}).Where("mlid IN (?)", deleteMlids).Delete(&models.FavourFoldersInfo{})
	}

	if len(updateList) > 0 {
		conf := config.GetConfig()
		for _, updateData := range updateList {
			existInfo := existMap[updateData.Mlid]
			oldTitle := strings.ReplaceAll(existInfo.Title, "/", "⁄")
			newTitle := strings.ReplaceAll(updateData.Title, "/", "⁄")
			if newTitle != oldTitle {
				oldPath := filepath.Join(
					conf.Download.Path,
					strconv.Itoa(existInfo.Mid),
					oldTitle,
				)
				newPath := filepath.Join(
					conf.Download.Path,
					strconv.Itoa(updateData.Mid),
					newTitle,
				)
				if _, err := os.Stat(oldPath); os.IsExist(err) {
					os.MkdirAll(newPath, os.ModePerm)
					if f, err := os.ReadDir(oldPath); err == nil {
						for _, v := range f {
							os.Rename(filepath.Join(oldPath, v.Name()), filepath.Join(newPath, v.Name()))
						}
						os.Remove(oldPath)
					}
				}
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

func DelFavourInfoByMid(mid int) {
	db := models.GetDB()
	db.Where(models.FavourFoldersInfo{Mid: mid}).Delete(&models.FavourFoldersInfo{})
}

type FavourFolders struct {
	Mlid       int    `json:"mlid"`
	Fid        int    `json:"fid"`
	Title      string `json:"title"`
	MediaCount int    `json:"media_count"`
	Sync       int    `json:"sync"`
}

func GetAccountFavourInfoByMid(mid int) *[]*FavourFolders {
	db := models.GetDB()
	var favourFolderInfos []models.FavourFoldersInfo
	db.Model(&models.FavourFoldersInfo{}).Where("mid = ?", mid).Find(&favourFolderInfos)
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
		db.Model(&models.FavourVideos{}).Where("mid = ? AND mlid = ? AND status = ?", mid, mlid, consts.VIDEO_STATUS_INIT).Update("status", consts.VIDEO_STATUS_TO_BE_DOWNLOAD)
	}
}

type FavFile struct {
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

func GetFavourIndex(mid, action, path string) map[string]interface{} {
	result := make(map[string]interface{})
	conf := config.GetConfig()
	rootPath := filepath.Join(conf.Download.Path, mid)

	result["adapter"] = mid
	result["dirname"] = path
	result["storages"] = []string{mid}

	subPath := filepath.Join(rootPath, strings.ReplaceAll(path, mid+"://", "/"))
	fileMap := make(map[string]*FavFile)
	fileNames := make([]string, 0)
	dirFiles, err := os.ReadDir(subPath)
	if err != nil {
		return result
	}

	for _, file := range dirFiles {
		if fileInfo, err := file.Info(); err == nil {
			file := FavFile{
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

	files := make([]*FavFile, 0)

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

// func GetFavourFilePreview(mid, action, path string) (string, error) {
// 	filePath, err := GetFavourFileDownload(mid, action, path)
// 	if action == "download" {
// 		return filePath, err
// 	}
// 	mtype, err := mimetype.DetectFile(filePath)
// 	if err != nil {
// 		return filePath, err
// 	}
// 	if strings.Contains(mtype.String(), "video") {
// 		if _, err = os.Stat(filePath + ".png"); err == nil {
// 			return filePath + ".png", nil
// 		} else {
// 			return "default_video_cover.png", nil
// 		}
// 	}
// 	return filePath, err

// }

func GetFavourFileDownload(mid, action, path string) (string, error) {
	conf := config.GetConfig()
	rootPath := filepath.Join(conf.Download.Path, mid)
	filePath := filepath.Join(rootPath, strings.ReplaceAll(path, mid+"://", "/"))
	if _, err := os.Stat(filePath); err != nil {
		return "", err
	} else {
		return filePath, nil
	}
}
