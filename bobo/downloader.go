package bobo

import (
	"bilibo/bobo/client"
	"bilibo/config"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/models"
	"bilibo/services"
	"bilibo/utils"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func downloadHandler(c *client.Client, video *models.FavourVideos, basePath, path string) {
	mid := c.GetMid()
	services.SetVideoStatus(video.ID, consts.VIDEO_STATUS_DOWNLOADING)

	videoStatus := consts.VIDEO_STATUS_DOWNLOAD_RETRY
	if fav := services.GetFavourInfoByMlid(video.Mlid); fav != nil {
		tmpFilePath := filepath.Join(basePath, ".tmp")
		fileName := fmt.Sprintf("%d_%d_%s_%d", mid, video.Mlid, video.Bvid, video.Cid)
		if dFilePath, dmimeType, err := c.DownloadVideoBestByBvidCid(
			video.Cid, video.Bvid, tmpFilePath, fileName,
		); err == nil {
			videoStatus = consts.VIDEO_STATUS_DOWNLOAD_DONE
			os.MkdirAll(path, os.ModePerm)
			distPath := filepath.Join(path,
				fmt.Sprintf("P%d %s.%s", video.Page, strings.ReplaceAll(video.Part, "/", "⁄"), dmimeType))
			utils.RenameDir(dFilePath, distPath)
		} else if err == consts.ERROR_DOWNLOAD_403 {
			errorInfo := fmt.Sprintf("user [%d] download video [%s] error: %v. try it later", mid, video.Bvid, err)
			services.SetVideoErrorMessage(video.Mlid, mid, video.Bvid, errorInfo)
			videoStatus = consts.VIDEO_STATUS_DOWNLOAD_RETRY
		} else {
			errorInfo := fmt.Sprintf("user [%d] get video [%s] info error: %v", mid, video.Bvid, err)
			services.SetVideoErrorMessage(video.Mlid, mid, video.Bvid, errorInfo)
		}
	} else {
		errorInfo := fmt.Sprintf("user [%d] video [%s] favour [%d] info not found in db", mid, video.Bvid, video.Mlid)
		services.SetVideoErrorMessage(video.Mlid, mid, video.Bvid, errorInfo)
	}
	services.SetVideoStatus(video.ID, videoStatus)
}

func downloadFavVideo(c *client.Client, ctx context.Context) {
	logger := log.GetLogger()
	mid := c.GetMid()
	conf := config.GetConfig()
	for {
		select {
		case <-ctx.Done():
			logger.Infof("user [%d] download exit", mid)
		default:
			logger.Infof("user [%d] download start", mid)
			video1 := services.GetToBeDownloadByMid(mid)
			video2 := services.GetRetryByMid(mid)
			if video1 == nil && video2 == nil {
				logger.Infof("user [%d] download finish. wait 4minutes", mid)
				time.Sleep(240 * time.Second)
				continue
			}
			if video1 != nil {
				if fav := services.GetFavourInfoByMlid(video1.Mlid); fav != nil {
					pathDst := filepath.Join(
						utils.GetFavourPath(mid, conf.Download.Path),
						strings.ReplaceAll(fav.Title, "/", "⁄"),
						strings.ReplaceAll(video1.Title, "/", "⁄"),
					)
					downloadHandler(c, video1, conf.Download.Path, pathDst)
				}
			}
			if video2 != nil {
				if fav := services.GetFavourInfoByMlid(video1.Mlid); fav != nil {
					pathDst := filepath.Join(
						utils.GetFavourPath(mid, conf.Download.Path),
						strings.ReplaceAll(fav.Title, "/", "⁄"),
						strings.ReplaceAll(video1.Title, "/", "⁄"),
					)
					downloadHandler(c, video2, conf.Download.Path, pathDst)
				}
			}
		}
		logger.Infof("user [%d] download end", mid)
	}
}
