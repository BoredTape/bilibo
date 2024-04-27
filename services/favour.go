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
)

func SetFavourInfo(favInfo *bili_client.AllFavourFolderInfo) {
	if favInfo == nil {
		return
	}
	for _, fav := range favInfo.List {
		f := models.FavourFoldersInfo{}
		db := models.GetDB()
		db.Where(models.FavourFoldersInfo{
			Mid: fav.Mid, Fid: fav.Fid,
		}).FirstOrInit(&f)
		needUpdata := false
		if f.ID == 0 {
			f.Sync = consts.FAVOUR_NOT_SYNC
			needUpdata = true
		}
		if f.Mid != fav.Mid {
			f.Mid = fav.Mid
			needUpdata = true
		}
		if f.Fid != fav.Fid {
			f.Fid = fav.Fid
			needUpdata = true
		}

		if f.Mlid != fav.Id {
			f.Mlid = fav.Id
			needUpdata = true
		}

		if f.Attr != fav.Attr {
			f.Attr = fav.Attr
			needUpdata = true
		}

		if f.FavState != fav.FavState {
			f.FavState = fav.FavState
			needUpdata = true
		}

		if f.MediaCount != fav.MediaCount {
			f.MediaCount = fav.MediaCount
			needUpdata = true
		}

		if f.ID > 0 && f.Title != fav.Title && f.Sync == consts.FAVOUR_NEED_SYNC {
			conf := config.GetConfig()
			oldPath := filepath.Join(
				conf.Download.Path,
				strconv.Itoa(f.Mid), strings.ReplaceAll(f.Title, "/", "⁄"))
			newPath := filepath.Join(
				conf.Download.Path,
				strconv.Itoa(f.Mid), strings.ReplaceAll(fav.Title, "/", "⁄"))
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

		if f.Title != fav.Title {
			f.Title = fav.Title
			needUpdata = true
		}
		if needUpdata {
			db.Save(&f)
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
