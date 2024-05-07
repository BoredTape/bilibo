package bobo

import (
	"bilibo/bobo/client"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/services"
	"fmt"
)

func (b *BoBo) RefreshFav(mid int) *client.AllFavourFolderInfo {
	logger := log.GetLogger()
	if client, err := b.GetClient(mid); err == nil {
		if data, err := client.GetAllFavourFolderInfo(mid, 2, 0); err == nil {
			folderInfo := make([]services.FolderInfo, 0)
			for _, v := range data.List {
				folderInfo = append(folderInfo, services.FolderInfo{
					Id:         v.Id,
					Fid:        v.Fid,
					Mid:        v.Mid,
					Attr:       v.Attr,
					Title:      v.Title,
					FavState:   v.FavState,
					MediaCount: v.MediaCount,
				})
			}
			serviceData := services.FavourFolderInfo{
				Count: data.Count,
				List:  folderInfo,
			}
			services.SetFavourInfo(mid, &serviceData)
			return data
		} else {
			logger.Warnf("client %d get fav list error: %v", mid, err)
		}
	}
	return nil
}

func (b *BoBo) RefreshFavVideo(mid int, data *client.AllFavourFolderInfo) map[string]*services.VideoInfo {
	logger := log.GetLogger()
	logger.Infof("user: %d refresh fav list", mid)
	videosInfoMap := make(map[string]*services.VideoInfo)
	if client, err := b.GetClient(mid); err == nil {
		if data != nil {
			for _, fav := range data.List {
				videosMap := make(map[string]*services.Video)
				mlid := fav.Id
				if fret, err := client.GetFavourList(mlid, 0, "", "", 0, 20, 1, "web"); err == nil {
					for _, media := range fret.Medias {
						if vret, err := client.GetVideoInfoByBvid(media.BvId); err == nil {
							for _, page := range vret.Pages {
								videosMapKey := fmt.Sprintf("%d_%s_%d", mlid, media.BvId, page.Cid)
								videosMap[videosMapKey] = &services.Video{
									Bvid:     media.BvId,
									SourceId: mlid,
									Mid:      mid,
									Cid:      page.Cid,
									Type:     consts.VIDEO_TYPE_FAVOUR,
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
						}
					}
				}
				if len(videosMap) > 0 {
					services.SetVideos(mid, mlid, videosMap, consts.VIDEO_TYPE_FAVOUR)
				}
			}
		} else {
			logger.Warnf("user %d get fav list empty", mid)
		}
	}
	return videosInfoMap
}
