package scheduler

import (
	"bilibo/bobo"

	"github.com/robfig/cron/v3"
)

var sched *cron.Cron

func GetCron() *cron.Cron {
	return sched
}

func init() {
	sched = cron.New()
}

func BiliBoSched(bobo *bobo.BoBo) {
	refreshWbiKey := refreshWbiKeyJob{"RefreshWbiKey", bobo}
	refreshWbiKey.Run()
	sched.AddJob("*/15 * * * *", &refreshWbiKey)

	refreshFavList := refreshFavListJob{"RefreshFavListJob", bobo}
	sched.AddJob("*/15 * * * *", &refreshFavList)

	refreshToView := refreshToViewJob{"RefreshToViewJob", bobo}
	sched.AddJob("*/15 * * * *", &refreshToView)
}

func Start() {
	sched.Start()
	t := InitTaskInfo("RefreshFavListJob", "更新收藏夹信息")
	t.UpdateNextRunningAt(15 * 60)
	t = InitTaskInfo("RefreshToViewJob", "更新稍后再看视频")
	t.UpdateNextRunningAt(15 * 60)
}

func DelJob(jobId int) {
	sched.Remove(cron.EntryID(jobId))
}
