package models

import (
	"time"

	"gorm.io/gorm"
)

type BiliAccounts struct {
	gorm.Model
	Cookies string
	Mid     int
	UName   string `gorm:"column:uname"`
	Face    string
	ImgKey  string
	SubKey  string
	Status  int
}

func (ba *BiliAccounts) TableName() string {
	return "bili_accounts"
}

type FavourFoldersInfo struct {
	gorm.Model
	Mlid       int    // 收藏夹mlid（完整id），收藏夹原始id+创建者mid尾号2位
	Fid        int    // 收藏夹原始id
	Mid        int    // 创建者mid
	Attr       int    // 属性位（？）
	Title      string // 收藏夹标题
	FavState   int    // 目标id是否存在于该收藏夹，存在于该收藏夹：1，不存在于该收藏夹：0
	MediaCount int    // 收藏夹内容数量
	Sync       int    // 是否同步，0：否，1：是
}

func (f *FavourFoldersInfo) TableName() string {
	return "favour_folders_info"
}

type Videos struct {
	gorm.Model
	Mlid   int
	Mid    int
	Bvid   string
	Cid    int
	Page   int
	Title  string
	Part   string
	Width  int // 当前分P 宽度
	Height int // 当前分P 高度
	Rotate int // 是否将宽高对换，0：正常，1：对换

	Status         int
	Type           int
	LastDownloadAt *time.Time
}

func (f *Videos) TableName() string {
	return "videos"
}

type QRCode struct {
	gorm.Model
	QRID   string `gorm:"column:qr_id"`
	Status int
}

func (f *QRCode) TableName() string {
	return "qr_code"
}

type VideoDownloadMessage struct {
	gorm.Model
	Mlid    int
	Mid     int
	Bvid    string
	Message string
	Type    int
}

func (f *VideoDownloadMessage) TableName() string {
	return "video_download_message"
}

type Task struct {
	gorm.Model
	TaskId        string `gorm:"column:task_id"`
	Name          string
	LastRunningAt *time.Time
	NextRunningAt *time.Time
	Type          int
}

func (f *Task) TableName() string {
	return "task"
}

type WatchLater struct {
	gorm.Model
	Mid  int
	Sync int
}

func (f *WatchLater) TableName() string {
	return "watch_later"
}
