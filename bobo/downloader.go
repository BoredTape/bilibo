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

func downloadHandler(c *client.Client, video *models.Videos, basePath, path string) {
	mid := c.GetMid()
	services.SetVideoStatus(video.ID, consts.VIDEO_STATUS_DOWNLOADING)

	videoStatus := consts.VIDEO_STATUS_DOWNLOAD_RETRY
	if fav := services.GetFavourInfoByMlid(video.SourceId); fav != nil {
		tmpFilePath := filepath.Join(basePath, ".tmp")
		fileName := fmt.Sprintf("%d_%d_%s_%d", mid, video.SourceId, video.Bvid, video.Cid)
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
			services.SetVideoErrorMessage(video.SourceId, mid, video.Bvid, errorInfo)
			videoStatus = consts.VIDEO_STATUS_DOWNLOAD_RETRY
		} else {
			errorInfo := fmt.Sprintf("user [%d] get video [%s] info error: %v", mid, video.Bvid, err)
			services.SetVideoErrorMessage(video.SourceId, mid, video.Bvid, errorInfo)
		}
	} else {
		errorInfo := fmt.Sprintf("user [%d] video [%s] favour [%d] info not found in db", mid, video.Bvid, video.SourceId)
		services.SetVideoErrorMessage(video.SourceId, mid, video.Bvid, errorInfo)
	}
	services.SetVideoStatus(video.ID, videoStatus)
}

func downloadVideo(c *client.Client, ctx context.Context) {
	logger := log.GetLogger()
	mid := c.GetMid()
	conf := config.GetConfig()
	for {
		accountInfo := services.GetAccountByMid(mid)
		t := services.NewTask(
			services.WithTaskType(consts.TASK_TYPE_RUNNING_TIME),
			services.WithName(fmt.Sprintf("用户 [%s] 的定时下载", accountInfo.UName)),
			services.WithTaskId(fmt.Sprintf("user_download_%d", mid)),
		)
		t.Save()
		select {
		case <-ctx.Done():
			logger.Infof("user [%d] download exit", mid)
		default:
			logger.Infof("user [%d] download start", mid)
			video1 := services.GetToBeDownloadByMid(mid)
			video2 := services.GetRetryByMid(mid)
			if video1 == nil && video2 == nil {
				logger.Infof("user [%d] download finish. wait 4minutes", mid)
				t.UpdateNextRunningAt(4 * 60)
				time.Sleep(240 * time.Second)
				continue
			}
			if video1 != nil {
				pathDst := ""
				if video1.Type == consts.VIDEO_TYPE_FAVOUR {
					if fav := services.GetFavourInfoByMlid(video1.SourceId); fav != nil {
						pathDst = filepath.Join(
							utils.GetFavourPath(mid, conf.Download.Path),
							strings.ReplaceAll(fav.Title, "/", "⁄"),
							strings.ReplaceAll(video1.Title, "/", "⁄"),
						)
					}
				} else if video1.Type == consts.VIDEO_TYPE_WATCH_LATER {
					pathDst = filepath.Join(
						utils.GetWatchLaterPath(mid, conf.Download.Path),
						strings.ReplaceAll(video1.Title, "/", "⁄"),
					)
				}

				if len(pathDst) > 0 {
					downloadHandler(c, video1, conf.Download.Path, pathDst)
				}

			}
			if video2 != nil {
				pathDst := ""
				if video2.Type == consts.VIDEO_TYPE_FAVOUR {
					if fav := services.GetFavourInfoByMlid(video2.SourceId); fav != nil {
						pathDst = filepath.Join(
							utils.GetFavourPath(mid, conf.Download.Path),
							strings.ReplaceAll(fav.Title, "/", "⁄"),
							strings.ReplaceAll(video2.Title, "/", "⁄"),
						)

					}
				} else if video2.Type == consts.VIDEO_TYPE_WATCH_LATER {
					pathDst = filepath.Join(
						utils.GetWatchLaterPath(mid, conf.Download.Path),
						strings.ReplaceAll(video2.Title, "/", "⁄"),
					)
				}
				if len(pathDst) > 0 {
					downloadHandler(c, video2, conf.Download.Path, pathDst)
				}
			}
		}
		logger.Infof("user [%d] download end", mid)
		t.UpdateNextRunningAt(4 * 60)
	}
}
