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
	go downloadFavVideo(c, ctx)
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
