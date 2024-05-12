package bobo

import (
	"bilibo/bobo/client"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/services"
	"fmt"
	"slices"
)

func (b *BoBo) RefreshCollected(mid int) *client.CollectedInfo {
	logger := log.GetLogger()
	if client, err := b.GetClient(mid); err == nil {
		if data, err := client.GetCollected(mid); err == nil {
			collectedInfo := make([]services.Collected, 0)
			for _, v := range data.List {
				collectedInfo = append(collectedInfo, services.Collected{
					Id:         v.Id,
					Mid:        v.Mid,
					Attr:       v.Attr,
					Title:      v.Title,
					MediaCount: v.MediaCount,
				})
			}
			serviceData := services.CollectedInfo{
				Count: data.Count,
				List:  collectedInfo,
			}
			services.SetCollectedInfo(mid, &serviceData)
			return data
		} else {
			logger.Warnf("client %d get fav list error: %v", mid, err)
		}
	}
	return nil
}

func (b *BoBo) RefreshCollectedVideo(mid int, data *client.CollectedInfo) map[string]*services.VideoInfo {
	logger := log.GetLogger()
	logger.Infof("user: %d collected video list", mid)
	videosInfoMap := make(map[string]*services.VideoInfo)

	if client, err := b.GetClient(mid); err == nil {
		if data != nil {
			for _, collected := range data.List {
				videosMap := make(map[string]*services.Video)
				invalidVideosBvidList := make([]string, 0)
				if fret, err := client.GetCollectedVideoList(collected.Id); err == nil {
					for _, media := range fret.Medias {
						if vret, code, err := client.GetVideoInfoByBvidCode(media.BvId); err == nil {
							for _, page := range vret.Pages {
								videosMapKey := fmt.Sprintf("%d_%s_%d", collected.Id, media.BvId, page.Cid)
								videosMap[videosMapKey] = &services.Video{
									SourceId: collected.Id,
									Bvid:     media.BvId,
									Cid:      page.Cid,
									Mid:      mid,
									Type:     consts.VIDEO_TYPE_COLLECTED,
								}

								videosInfoMapKey := fmt.Sprintf("%s_%d", media.BvId, page.Cid)
								videosInfoMap[videosInfoMapKey] = &services.VideoInfo{
									Bvid:   media.BvId,
									Cid:    page.Cid,
									Page:   page.Page,
									Title:  vret.Title,
									Part:   page.Part,
									Width:  vret.Dimension.Width,
									Height: vret.Dimension.Height,
									Rotate: vret.Dimension.Rotate,
								}
							}
						} else if code == 62002 {
							logger.Infof("用户: %d 收藏和订阅: %s 无效视频bvid: %s", mid, collected.Title, media.BvId)
							if !slices.Contains(invalidVideosBvidList, media.BvId) {
								invalidVideosBvidList = append(invalidVideosBvidList, media.BvId)
							}
						}
					}
				}
				if len(invalidVideosBvidList) > 0 {
					logger.Infof("用户: %d 收藏和订阅: %s 有 %d 个无效视频", mid, collected.Title, len(invalidVideosBvidList))
					services.SetInvalidVideos(mid, collected.Id, invalidVideosBvidList, consts.VIDEO_TYPE_COLLECTED)
				}
				if len(videosMap) > 0 {
					services.SetVideos(mid, collected.Id, videosMap, consts.VIDEO_TYPE_COLLECTED)
				}
			}
		} else {
			logger.Warnf("user %d get fav list empty", mid)
		}
	}
	return videosInfoMap
}
