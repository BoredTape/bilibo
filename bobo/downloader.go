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
	"time"
)

func downloadHandler(c *client.Client, video *models.Videos, basePath, path string) {
	mid := c.GetMid()
	services.SetVideoStatus(video.ID, consts.VIDEO_STATUS_DOWNLOADING)
	videoStatus := consts.VIDEO_STATUS_DOWNLOAD_RETRY

	videoInfo := services.GetVideoInfo(video.Bvid, video.Cid)
	if videoInfo == nil {
		services.SetVideoStatus(video.ID, videoStatus)
		return
	}

	tmpFilePath := filepath.Join(basePath, ".tmp")
	fileName := fmt.Sprintf("%d_%d_%s_%d", mid, video.SourceId, video.Bvid, video.Cid)
	if dFilePath, dmimeType, err := c.DownloadVideoBestByBvidCid(
		video.Cid, video.Bvid, tmpFilePath, fileName,
	); err == nil {
		videoStatus = consts.VIDEO_STATUS_DOWNLOAD_DONE
		os.MkdirAll(path, os.ModePerm)
		distPath := filepath.Join(path,
			fmt.Sprintf("P%d %s.%s", videoInfo.Page, utils.Name(videoInfo.Part), dmimeType))
		utils.RenameDir(dFilePath, distPath)
	} else if err == consts.ERROR_DOWNLOAD_403 {
		errorInfo := fmt.Sprintf("user [%d] download video [%s] error: %v. try it later", mid, video.Bvid, err)
		services.SetVideoErrorMessage(video.SourceId, mid, video.Type, video.Bvid, errorInfo)
		videoStatus = consts.VIDEO_STATUS_DOWNLOAD_RETRY
	} else {
		errorInfo := fmt.Sprintf("user [%d] get video [%s] info error: %v", mid, video.Bvid, err)
		services.SetVideoErrorMessage(video.SourceId, mid, video.Type, video.Bvid, errorInfo)
	}

	services.SetVideoStatus(video.ID, videoStatus)
}

func downloadVideo(c *client.Client, ctx context.Context) {
	logger := log.GetLogger()
	mid := c.GetMid()
	conf := config.GetConfig()
	accountInfo := services.GetAccountByMid(mid)
	t := services.NewTask(
		services.WithTaskType(consts.TASK_TYPE_RUNNING_TIME),
		services.WithName(fmt.Sprintf("用户 [%s] 的定时下载", accountInfo.UName)),
		services.WithTaskId(fmt.Sprintf("user_download_%d", mid)),
	)
	t.Save()
	for {
		select {
		case <-ctx.Done():
			logger.Infof("user [%d] download exit", mid)
			t.Delete()
			return
		default:
			logger.Infof("user [%d] download start", mid)
			video1 := services.GetToBeDownloadByMid(mid)
			video2 := services.GetRetryByMid(mid)
			if video1 == nil && video2 == nil {
				logger.Infof("user [%d] download finish. wait 4minutes", mid)
				t.UpdateNextRunningAt(4 * 60)
				time.Sleep(230 * time.Second)
				continue
			} else {
				if video1 != nil {
					video1Info := services.GetVideoInfo(video1.Bvid, video1.Cid)
					pathDst := ""
					if video1.Type == consts.VIDEO_TYPE_FAVOUR {
						if fav := services.GetFavourInfoByMlid(video1.SourceId); fav != nil {
							pathDst = filepath.Join(
								utils.GetFavourPath(mid, conf.Download.Path),
								utils.Name(fav.Title),
								utils.Name(video1Info.Title),
							)
						}
					} else if video1.Type == consts.VIDEO_TYPE_WATCH_LATER {
						pathDst = filepath.Join(
							utils.GetWatchLaterPath(mid, conf.Download.Path),
							utils.Name(video1Info.Title),
						)
					} else if video1.Type == consts.VIDEO_TYPE_COLLECTED {
						if collected := services.GetCollectedInfoByCollidMid(mid, video1.SourceId); collected != nil {
							pathDst = filepath.Join(
								utils.GetCollectedPath(mid, conf.Download.Path),
								utils.Name(collected.Title),
								utils.Name(video1Info.Title),
							)
						}
					} else {
						logger.Info("视频类型不支持：", video1.Type)
					}

					if len(pathDst) > 0 {
						downloadHandler(c, video1, conf.Download.Path, pathDst)
					}

				}
				if video2 != nil {
					video2Info := services.GetVideoInfo(video2.Bvid, video2.Cid)
					pathDst := ""
					if video2.Type == consts.VIDEO_TYPE_FAVOUR {
						if fav := services.GetFavourInfoByMlid(video2.SourceId); fav != nil {
							pathDst = filepath.Join(
								utils.GetFavourPath(mid, conf.Download.Path),
								utils.Name(fav.Title),
								utils.Name(video2Info.Title),
							)

						}
					} else if video2.Type == consts.VIDEO_TYPE_WATCH_LATER {
						pathDst = filepath.Join(
							utils.GetWatchLaterPath(mid, conf.Download.Path),
							utils.Name(video2Info.Title),
						)
					} else if video2.Type == consts.VIDEO_TYPE_COLLECTED {
						if collected := services.GetCollectedInfoByCollidMid(mid, video2.SourceId); collected != nil {
							pathDst = filepath.Join(
								utils.GetCollectedPath(mid, conf.Download.Path),
								utils.Name(collected.Title),
								utils.Name(video2Info.Title),
							)
						}
					} else {
						logger.Info("视频类型不支持：", video2.Type)
					}

					if len(pathDst) > 0 {
						downloadHandler(c, video2, conf.Download.Path, pathDst)
					}
				}
			}
			logger.Infof("user [%d] download end", mid)
			t.UpdateNextRunningAt(4 * 60)
		}
		time.Sleep(10 * time.Second)
	}
}
