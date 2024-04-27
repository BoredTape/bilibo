package services

import (
	"bilibo/models"
)

func SetTask(mid int, jobId int, taskType int) {
	task := models.Tasks{}
	db := models.GetDB()
	db.Where(models.Tasks{Mid: mid, Type: taskType}).FirstOrInit(&task)
	task.JobId = jobId
	db.Save(&task)
}

func DelTaskByMid(mid int) []int {
	db := models.GetDB()
	var tasks []models.Tasks
	db.Where(models.Tasks{Mid: mid}).Find(&tasks)
	jobIds := make([]int, 0)
	for _, task := range tasks {
		db.Delete(&task)
		jobIds = append(jobIds, task.JobId)
	}
	return jobIds
}
