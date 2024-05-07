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

func handleClient() {
	for ch := range *universal.GetCH() {
		if ch.Action == consts.CHANNEL_ACTION_ADD_CLIENT {
			addClient(&ch, true)
		} else if ch.Action == consts.CHANNEL_ACTION_DELETE_CLIENT {
			delClient(&ch)
		} else {
			logger := log.GetLogger()
			logger.Info(fmt.Sprintf("channel get unknown action: %d", ch.Action))
		}
	}
}

func addClient(ch *universal.CH, init bool) {
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
	if init {
		go bobo.RefreshAll(ch.Mid)
	}
	go downloadVideo(c, ctx)
}

func (b *BoBo) RefreshAll(mid int) {
	logger := log.GetLogger()
	videosInfoMap := make(map[string]*services.VideoInfo)

	favData := bobo.RefreshFav(mid)
	favVideosInfoMap := bobo.RefreshFavVideo(mid, favData)
	logger.Infof("收藏夹的视频数有 %d 个", len(favVideosInfoMap))
	if len(favVideosInfoMap) > 0 {
		for k, v := range favVideosInfoMap {
			videosInfoMap[k] = v
		}
	}

	twVideosInfoMap := bobo.RefreshToView(mid)
	logger.Infof("稍后再看的视频数有 %d 个", len(twVideosInfoMap))
	if len(twVideosInfoMap) > 0 {
		for k, v := range twVideosInfoMap {
			videosInfoMap[k] = v
		}
	}

	collectedData := bobo.RefreshCollected(mid)
	cVideosInfoMap := bobo.RefreshCollectedVideo(mid, collectedData)
	logger.Infof("收藏和订阅的视频数有 %d 个", len(cVideosInfoMap))
	if len(cVideosInfoMap) > 0 {
		for k, v := range cVideosInfoMap {
			videosInfoMap[k] = v
		}
	}

	if len(videosInfoMap) > 0 {
		logger.Infof("受到影响的视频数有 %d 个", len(videosInfoMap))
		services.SetVideosInfo(videosInfoMap)
	}
	services.AfterRefresh(mid)
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
		}, false)
	}
}
