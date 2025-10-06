package controllers

import (
	"kanban/config"
	"kanban/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateTask(c *gin.Context) {
	var input struct {
		Title     string `json:"title"`
		Status    string `json:"status"`
		SprintID  uint   `json:"sprint_id"`
		AssignTo  uint   `json:"assign_to"`
		Estimation float64 `json:"estimation"`		
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title == "" || input.Status == "" || input.SprintID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	task := models.Task{
		Title:      input.Title,
		Status:     input.Status,
		SprintID:   input.SprintID,
		Estimation: input.Estimation,
	}

	if input.AssignTo != 0 {
		task.AssignTo = &input.AssignTo
	}

	if err := config.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": task})
}

func GetAllTasks(c *gin.Context) {
	var tasks []models.Task
	if err := config.DB.Preload("Sprint").Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tasks})
}

func GetTasks(c *gin.Context) {
	id := c.Param("id")

	var tasks []models.Task
	if err := config.DB.Where("sprint_id = ?", id).Preload("Sprint").Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tasks})
}

func UpdateTaskStatus(c *gin.Context) {
	id := c.Param("id")
	var task models.Task

	if err := config.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task tidak ditemukan"})
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	c.BindJSON(&body)

	task.Status = body.Status
	config.DB.Save(&task)

	c.JSON(http.StatusOK, task)
}

func AssignToUser(c *gin.Context) {
	id := c.Param("id")
	var task models.Task
	if err := config.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	var body struct {
		AssignTo uint `json:"assign_to"`
	}
	c.BindJSON(&body)
	
	// If AssignTo is 0, set to nil (unassign)
	if body.AssignTo == 0 {
		task.AssignTo = nil
	} else {
		task.AssignTo = &body.AssignTo
	}
	
	config.DB.Save(&task)

	c.JSON(http.StatusOK, task)
}

func DeleteTask(c *gin.Context) {
	id := c.Param("id")
	var task models.Task
	if err := config.DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	config.DB.Delete(&task)
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}