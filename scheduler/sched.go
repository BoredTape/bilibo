package scheduler

import (
	"bilibo/bili"

	"github.com/robfig/cron/v3"
)

var sched *cron.Cron

func GetCron() *cron.Cron {
	return sched
}

func init() {
	sched = cron.New()
}

func BiliBoSched(bobo *bili.BoBo) {
	refreshWbiKey := refreshWbiKeyJob{bobo}
	refreshWbiKey.Run()
	sched.AddJob("*/15 * * * *", &refreshWbiKey)

	refreshFavList := refreshFavListJob{bobo}
	refreshFavList.SetFav()
	sched.AddJob("*/15 * * * *", &refreshFavList)
}

func Start() {
	sched.Start()
}

func DelJob(jobId int) {
	sched.Remove(cron.EntryID(jobId))
}
