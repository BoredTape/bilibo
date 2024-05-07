package services

import (
	"bilibo/models"
	"time"
)

type TaskInfo struct {
	id            uint
	taskId        string
	name          string
	lastRunningAt *time.Time
	nextRunningAt *time.Time
	taskType      int
}

type TaskInfoOption func(t *TaskInfo)

func WithTaskId(taskId string) TaskInfoOption {
	return func(t *TaskInfo) {
		t.taskId = taskId
	}
}

func WithName(name string) TaskInfoOption {
	return func(t *TaskInfo) {
		t.name = name
	}
}

func WithTaskType(taskType int) TaskInfoOption {
	return func(t *TaskInfo) {
		t.taskType = taskType
	}
}

func NewTask(opts ...TaskInfoOption) *TaskInfo {
	t := &TaskInfo{lastRunningAt: nil}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t *TaskInfo) Save() {
	timeNow := time.Now()
	db := models.GetDB()
	task := &models.Task{}
	db.Model(&models.Task{}).Where("task_id = ?", t.taskId).FirstOrInit(&task)
	task.TaskId = t.taskId
	task.Name = t.name
	task.LastRunningAt = &timeNow
	task.NextRunningAt = t.nextRunningAt
	task.Type = t.taskType
	if t.id > 0 {
		task.ID = t.id
	}
	db.Save(&task)
	t.id = task.ID
}

func (t *TaskInfo) UpdateNextRunningAt(seconds int) {
	nextTime := time.Now().Add(time.Second * time.Duration(seconds))
	t.nextRunningAt = &nextTime
	models.GetDB().Model(
		&models.Task{},
	).Select("next_running_at").Where("id = ?", t.id).Updates(
		&models.Task{
			NextRunningAt: &nextTime,
		},
	)
}

func (t *TaskInfo) Delete() {
	db := models.GetDB()
	db.Delete(&models.Task{}, t.id)
}
