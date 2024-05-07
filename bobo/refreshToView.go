package bobo

import (
	"bilibo/consts"
	"bilibo/services"
	"fmt"
)

func (b *BoBo) RefreshToView(mid int) map[string]*services.VideoInfo {
	videosMap := make(map[string]*services.Video)
	videosInfoMap := make(map[string]*services.VideoInfo)
	if client, err := b.GetClient(mid); err == nil {
		if toViewData, err := client.GetToView(); err == nil {
			for _, data := range toViewData.List {
				if vret, err := client.GetVideoInfoByBvid(data.Bvid); err == nil {
					for _, page := range vret.Pages {
						videosMapKey := fmt.Sprintf("%d_%s_%d", 0, data.Bvid, page.Cid)
						videosMap[videosMapKey] = &services.Video{
							Bvid:     data.Bvid,
							SourceId: 0,
							Mid:      mid,
							Cid:      page.Cid,
							Type:     consts.VIDEO_TYPE_WATCH_LATER,
						}
						videosInfoMapKey := fmt.Sprintf("%s_%d", data.Bvid, page.Cid)
						videosInfoMap[videosInfoMapKey] = &services.VideoInfo{
							Bvid:   data.Bvid,
							Cid:    page.Cid,
							Page:   page.Page,
							Title:  vret.Title,
							Part:   page.Part,
							Width:  vret.Dimension.Width,
							Height: vret.Dimension.Height,
							Rotate: vret.Dimension.Rotate,
						}
					}
				}
			}
		}
	}
	if len(videosMap) > 0 {
		services.SetVideos(mid, 0, videosMap, consts.VIDEO_TYPE_WATCH_LATER)
	}
	return videosInfoMap
}
