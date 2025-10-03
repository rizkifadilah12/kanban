package controllers

import (
	"net/http"
	"kanban/config"
	"kanban/models"

	"github.com/gin-gonic/gin"
)

func CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Create(&task)
	c.JSON(http.StatusOK, task)
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
