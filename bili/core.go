package bili

import (
	"bilibo/bili/bili_client"
	"bilibo/models"
	"context"
	"errors"
	"time"
)

var biliBo *BoBo

type BoBo struct {
	client           map[int]*bili_client.Client
	clientCancelFunc map[int]context.CancelFunc
}

func InitBiliBo() {
	biliBo = New()
}
func GetBilibo() *BoBo {
	return biliBo
}

func New() *BoBo {
	b := &BoBo{
		client:           make(map[int]*bili_client.Client),
		clientCancelFunc: make(map[int]context.CancelFunc),
	}
	b.restore()
	return b
}
func NewClient(ua string, timeout time.Duration) *bili_client.Client {
	return bili_client.NewClient(
		bili_client.WithUA(ua),
		bili_client.WithTimeout(timeout),
	)
}

func (b *BoBo) restore() {
	var accounts []models.BiliAccounts
	models.GetDB().Model(&models.BiliAccounts{}).Find(&accounts)

	for _, account := range accounts {
		c := bili_client.NewClient(
			bili_client.WithMid(account.Mid),
			bili_client.WithCookiesStrings(account.Cookies),
			bili_client.WithImgKey(account.ImgKey),
			bili_client.WithSubKey(account.SubKey),
		)
		b.client[account.Mid] = c
	}
}

func (b *BoBo) AddClient(c *bili_client.Client) (int, int64, error) {
	nav, errorCode, err := c.GetNavigation()
	if err != nil {
		return 0, errorCode, err
	}
	if err := c.RefreshWbiKey(nav); err == nil {
		if mid := c.GetMid(); mid != 0 {
			b.client[mid] = c
			return mid, 0, nil
		} else {
			return 0, 0, errors.New("get mid error")
		}
	} else {
		return 0, errorCode, err
	}
}

func (b *BoBo) DelClient(mid int) {
	delete(b.client, mid)
	if cancelFunc, ok := b.clientCancelFunc[mid]; ok {
		cancelFunc()
	}
	delete(b.clientCancelFunc, mid)
}

func (b *BoBo) GetClient(mid int) (*bili_client.Client, error) {
	if client, ok := b.client[mid]; ok {
		return client, nil
	}
	return nil, errors.New("client not found")
}

func (b *BoBo) ClientList() []int {
	var list []int
	for k := range b.client {
		list = append(list, k)
	}
	return list
}

func (b *BoBo) ClientSetCancal(mid int, cancelFunc context.CancelFunc) {
	if _, err := b.GetClient(mid); err == nil {
		if oldFunc, ok := b.clientCancelFunc[mid]; ok {
			oldFunc()
		}
		b.clientCancelFunc[mid] = cancelFunc
	}
}
