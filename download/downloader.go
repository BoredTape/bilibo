package download

import (
	"bilibo/bili"
	"bilibo/config"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/models"
	"bilibo/services"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func DownloadHandler(mid int, video *models.FavourVideos) {
	conf := config.GetConfig()
	biliBo := bili.GetBilibo()
	client, _ := biliBo.GetClient(mid)
	services.SetVideoStatus(video.ID, consts.VIDEO_STATUS_DOWNLOADING)

	videoStatus := consts.VIDEO_STATUS_DOWNLOAD_RETRY
	if fav := services.GetFavourInfoByMlid(video.Mlid); fav != nil {
		tmpFilePath := filepath.Join(conf.Download.Path, ".tmp")
		fileName := fmt.Sprintf("%d_%d_%s_%d", mid, video.Mlid, video.Bvid, video.Cid)
		if dFilePath, dmimeType, err := client.DownloadVideoBestByBvidCid(
			video.Cid, video.Bvid, tmpFilePath, fileName,
		); err == nil {
			videoStatus = consts.VIDEO_STATUS_DOWNLOAD_DONE

			pathDst := filepath.Join(
				conf.Download.Path,
				strconv.Itoa(video.Mid),
				strings.ReplaceAll(fav.Title, "/", "⁄"),
				strings.ReplaceAll(video.Title, "/", "⁄"),
			)
			os.MkdirAll(pathDst, os.ModePerm)
			distPath := filepath.Join(
				pathDst,
				fmt.Sprintf("P%d %s.%s", video.Page, strings.ReplaceAll(video.Part, "/", "⁄"), dmimeType))
			os.Rename(dFilePath, distPath)
			// bili_client.GenerateCover(distPath)
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

func AccountDownload(mid int, ctx context.Context) {
	logger := log.GetLogger()
	for {
		select {
		case <-ctx.Done():
			logger.Infof("user [%d] download exit", mid)
		default:
			logger.Infof("user [%d] download start", mid)
			biliBo := bili.GetBilibo()
			if _, err := biliBo.GetClient(mid); err == nil {
				video1 := services.GetToBeDownloadByMid(mid)
				video2 := services.GetRetryByMid(mid)
				if video1 == nil && video2 == nil {
					logger.Infof("user [%d] download finish. wait 4minutes", mid)
					time.Sleep(240 * time.Second)
					continue
				}
				if video1 != nil {
					DownloadHandler(mid, video1)
				}
				if video2 != nil {
					DownloadHandler(mid, video2)
				}
			} else {
				logger.Errorf("user [%d] get client error: %v", mid, err)
				services.SetUserVideosStatus(mid, consts.VIDEO_STATUS_DOWNLOAD_FAIL)
				break
			}
			logger.Infof("user [%d] download end", mid)
		}
	}
}
