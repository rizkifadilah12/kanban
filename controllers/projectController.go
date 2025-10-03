package controllers

import (
	"net/http"
	"kanban/config"
	"kanban/models"

	"github.com/gin-gonic/gin"
)

func CreateProject(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Create(&project)
	c.JSON(http.StatusOK, project)
}

func GetProjects(c *gin.Context) {
	var projects []models.Project
	config.DB.Preload("Tasks").Find(&projects)
	c.JSON(http.StatusOK, projects)
}
