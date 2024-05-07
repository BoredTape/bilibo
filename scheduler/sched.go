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

	refreshVideoList := refreshVideoJob{"refreshVideoJob", bobo}
	sched.AddJob("*/15 * * * *", &refreshVideoList)
}

func Start() {
	sched.Start()
	t := InitTaskInfo("refreshVideoJob", "更新用户视频信息")
	t.UpdateNextRunningAt(15 * 60)
}

func DelJob(jobId int) {
	sched.Remove(cron.EntryID(jobId))
}
