package utils

import (
	"bilibo/consts"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func Name(sourceName string) string {
	return strings.ReplaceAll(sourceName, "/", "⁄")
}

func getPath(mid int, basePath string, subPath string) string {
	return filepath.Join(
		basePath,
		strconv.Itoa(mid),
		subPath,
	)
}

func GetFavourPath(mid int, basePath string) string {
	return getPath(mid, basePath, consts.ACCOUNT_DIR_FAVOUR)
}

func GetWatchLaterPath(mid int, basePath string) string {
	return getPath(mid, basePath, consts.ACCOUNT_DIR_WATCH_LATER)
}

func GetRecyclePath(mid int, basePath string) string {
	return getPath(mid, basePath, consts.ACCOUNT_DIR_RECYCLE)
}

func InitAccountPath(mid int, basePath string) error {
	for _, dir := range consts.GET_ACCOUNT_DIR() {
		fullDir := getPath(mid, basePath, dir)
		if err := os.MkdirAll(fullDir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func RenameDir(oldPath string, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}

func RecyclePath(mid int, basePath, favourName string) error {
	recyclePath := GetRecyclePath(mid, basePath)
	timeNow := time.Now()
	favourPath := filepath.Join(
		GetFavourPath(mid, basePath), favourName,
	)
	deletePath := filepath.Join(recyclePath, fmt.Sprintf("%s_%d", favourName, timeNow.Unix()))
	return RenameDir(favourPath, deletePath)
}
