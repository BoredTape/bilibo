package services

import (
	"bilibo/models"
	"time"
)

type Task struct {
	TaskId        string     `json:"task_id"`
	Name          string     `json:"name"`
	LastRunningAt *time.Time `json:"last_running_at"`
	NextRunningAt *time.Time `json:"next_running_at"`
	Type          int        `json:"type"`
}

func TaskList() []*Task {
	db := models.GetDB()
	taskList := make([]*Task, 0)
	dbTaskList := make([]*models.Task, 0)
	db.Model(&models.Task{}).Order("next_running_at asc").Find(&dbTaskList)
	for _, v := range dbTaskList {
		taskList = append(taskList, &Task{
			TaskId:        v.TaskId,
			Name:          v.Name,
			LastRunningAt: v.LastRunningAt,
			NextRunningAt: v.NextRunningAt,
			Type:          v.Type,
		})
	}
	return taskList
}
