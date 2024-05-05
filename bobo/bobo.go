package bobo

import (
	"bilibo/bobo/client"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/services"
	"bilibo/universal"
	"context"
	"errors"
	"fmt"
)

var bobo *BoBo

type BoBo struct {
	client           map[int]*client.Client
	clientCancelFunc map[int]context.CancelFunc
}

func Init() {
	bobo = &BoBo{
		client:           make(map[int]*client.Client),
		clientCancelFunc: make(map[int]context.CancelFunc),
	}
	restoreClient()
	go handleClient()
}

func GetBoBo() *BoBo {
	return bobo
}

func (b *BoBo) ClientList() []int {
	var list []int
	for k := range b.client {
		list = append(list, k)
	}
	return list
}

func (b *BoBo) GetClient(mid int) (*client.Client, error) {
	if client, ok := b.client[mid]; ok {
		return client, nil
	}
	return nil, errors.New("client not found")
}

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

func (b *BoBo) RefreshFavVideo(mid int, data *client.AllFavourFolderInfo) {
	logger := log.GetLogger()
	logger.Infof("user: %d refresh fav list", mid)
	fv_svc := services.VideoService{}
	if client, err := b.GetClient(mid); err == nil {
		if data != nil {
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
								fv_svc.V.Type = consts.VIDEO_TYPE_FAVOUR
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

func (b *BoBo) RefreshToView(mid int) {
	fv_svc := services.VideoService{}
	if client, err := b.GetClient(mid); err == nil {
		fv_svc.SetMid(mid)
		if toViewData, err := client.GetToView(); err == nil {
			for _, data := range toViewData.List {
				bvid := data.Bvid
				fv_svc.V.Bvid = bvid
				fv_svc.V.Mlid = 0
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
						fv_svc.V.Type = consts.VIDEO_TYPE_WATCH_LATER
						fv_svc.Save()
					}
				}
			}
		}
	}
}

func handleClient() {
	for ch := range *universal.GetCH() {
		if ch.Action == consts.CHANNEL_ACTION_ADD_CLIENT {
			addClient(&ch)
		} else if ch.Action == consts.CHANNEL_ACTION_DELETE_CLIENT {
			delClient(&ch)
		} else {
			logger := log.GetLogger()
			logger.Info(fmt.Sprintf("channel get unknown action: %d", ch.Action))
		}
	}
}

func addClient(ch *universal.CH) {
	c := client.New(
		client.WithUA(""), // 先使用默认的，可能以后会改成可修改的
		client.WithCookiesStrings(ch.Cookies),
		client.WithImgKey(ch.ImgKey),
		client.WithSubKey(ch.SubKey),
		client.WithMid(ch.Mid),
	)
	bobo.client[ch.Mid] = c
	ctx, cancel := context.WithCancel(context.Background())
	bobo.clientCancelFunc[ch.Mid] = cancel
	go bobo.RefreshAll(ch.Mid)
	go downloadFavVideo(c, ctx)
}

func (b *BoBo) RefreshAll(mid int) {
	data := bobo.RefreshFav(mid)
	bobo.RefreshFavVideo(mid, data)
	bobo.RefreshToView(mid)
}

func (b *BoBo) DelClient(mid int) {
	delete(b.client, mid)
	if cancelFunc, ok := bobo.clientCancelFunc[mid]; ok {
		cancelFunc()
	}
	delete(b.clientCancelFunc, mid)
}

func delClient(ch *universal.CH) {
	bobo.DelClient(ch.Mid)
}

func restoreClient() {
	accounts := services.GetAccountList()
	for _, account := range *accounts {
		addClient(&universal.CH{
			Mid:     account.Mid,
			UName:   account.UName,
			Face:    account.Face,
			ImgKey:  account.ImgKey,
			SubKey:  account.SubKey,
			Cookies: account.Cookies,
			Action:  consts.CHANNEL_ACTION_ADD_CLIENT,
		})
	}
}
