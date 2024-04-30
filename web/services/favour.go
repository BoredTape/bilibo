package services

import (
	"bilibo/config"
	"bilibo/consts"
	"bilibo/models"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/maruel/natural"
)

func DelFavourInfoByMid(mid int) {
	db := models.GetDB()
	db.Where(models.FavourFoldersInfo{Mid: mid}).Delete(&models.FavourFoldersInfo{})
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
