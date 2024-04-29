package scheduler

import (
	"bilibo/bili"
	"bilibo/bili/bili_client"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/services"
)

type refreshWbiKeyJob struct {
	bobo *bili.BoBo
}

func (r *refreshWbiKeyJob) Run() {
	logger := log.GetLogger()
	logger.Info("refresh wbi key")
	for _, clientId := range r.bobo.ClientList() {
		if client, err := r.bobo.GetClient(clientId); err == nil {
			nav, errorCode, err := client.GetNavigation()
			if err != nil {
				logger.Error("client %d get nav error: %v", clientId, err)
			}
			if err := client.RefreshWbiKey(nav); err == nil {
				mid := client.GetMid()
				if mid == 0 {
					logger.Error("get mid error")
					continue
				}
				imgKey, subKey, _, err := client.GetWbi()
				if err != nil {
					logger.Error(err)
					continue
				}
				services.UpdateAccountWBI(
					mid, imgKey, subKey,
				)
			} else if errorCode == -101 {
				services.SetAccountStatus(client.GetMid(), consts.ACCOUNT_STATUS_NOT_LOGIN)
				r.bobo.DelClient(client.GetMid())
			} else if errorCode != 0 && errorCode != -101 {
				services.SetAccountStatus(client.GetMid(), consts.ACCOUNT_STATUS_INVALID)
				r.bobo.DelClient(client.GetMid())
			}
		}
	}

}

type refreshFavListJob struct {
	bobo *bili.BoBo
}

func (r *refreshFavListJob) Run() {
	logger := log.GetLogger()
	for _, mid := range r.bobo.ClientList() {
		if client, err := r.bobo.GetClient(mid); err == nil {
			logger.Infof("user: %d refresh fav list", mid)
			fv_svc := services.FavourVideoService{}
			if data := r.SetFav(); data != nil {
				fv_svc.SetMid(mid)
				for _, fav := range data.List {
					mlid := fav.Id
					fv_svc.V.Mlid = mlid
					if fret, err := client.GetFavourList(mlid, 0, "", "", 0, 20, 1, "web"); err == nil {
						for _, media := range fret.Medias {
							bvid := media.BvId
							fv_svc.V.Bvid = bvid
							if vret, err := client.GetVideoInfoByBvid(bvid); err == nil {
								for _, page := range vret.Pages {
									cid := page.Cid
									fv_svc.V.Cid = cid
									fv_svc.V.Title = vret.Title
									fv_svc.V.Part = page.Part
									fv_svc.V.Height = page.Dimension.Height
									fv_svc.V.Width = page.Dimension.Width
									fv_svc.V.Rotate = page.Dimension.Rotate
									fv_svc.V.Page = page.Page
									fv_svc.Save()
								}

							}
						}
					}

				}
			} else {
				logger.Warnf("user %d get fav list empty", mid)
			}
		}
	}
}

func (r *refreshFavListJob) SetFav() *bili_client.AllFavourFolderInfo {
	logger := log.GetLogger()
	for _, mid := range r.bobo.ClientList() {
		if client, err := r.bobo.GetClient(mid); err == nil {
			if data, err := client.GetAllFavourFolderInfo(mid, 2, 0); err == nil {
				services.SetFavourInfo(mid, data)
				return data
			} else {
				logger.Warnf("client %d get fav list error: %v", mid, err)
			}
		}
	}
	return nil
}
