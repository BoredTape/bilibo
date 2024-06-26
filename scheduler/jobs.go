package scheduler

import (
	"bilibo/bobo"
	"bilibo/consts"
	"bilibo/log"
	"bilibo/services"
)

type refreshWbiKeyJob struct {
	taskId string
	bobo   *bobo.BoBo
}

func (r *refreshWbiKeyJob) Run() {
	t := InitTaskInfo(r.taskId, "更新用户信息")
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
	t.UpdateNextRunningAt(15 * 60)
}

type refreshVideoJob struct {
	taskId string
	bobo   *bobo.BoBo
}

func (r *refreshVideoJob) Run() {
	t := InitTaskInfo(r.taskId, "更新用户视频信息")
	logger := log.GetLogger()
	logger.Info("refresh user video info")
	for _, clientId := range r.bobo.ClientList() {
		r.bobo.RefreshAll(clientId)
	}
	t.UpdateNextRunningAt(15 * 60)
}

func InitTaskInfo(taskId string, name string) *services.TaskInfo {
	t := services.NewTask(
		services.WithTaskId(taskId),
		services.WithName(name),
		services.WithTaskType(consts.TASK_TYPE_SCHEDULER),
	)
	t.Save()
	return t
}
