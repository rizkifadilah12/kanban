package controllers

import (
	"fmt"
	"kanban/config"
	"kanban/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateProject(c *gin.Context) {
	type CreateProjectInput struct {
		Name           string `json:"name"`
		Description    string `json:"description"`
		Hours          int    `json:"hours"`
		StoryPoints    int    `json:"story_points"`
		ParticipantIDs []uint `json:"participant_ids"`
	}

	var input CreateProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project := models.Project{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := config.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if input.Name == "" || input.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	if len(input.ParticipantIDs) > 0 {
		var participants []models.User
		if err := config.DB.Find(&participants, input.ParticipantIDs).Error; err == nil {
			config.DB.Model(&project).Association("UserParticipants").Append(participants)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"data": project})
}

func GetProjects(c *gin.Context) {
	id := c.Param("id")

	var project models.Project
	if err := config.DB.Preload("UserParticipants").First(&project, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": project})
}

func GetAllProjects(c *gin.Context) {
	var projects []models.Project
	if err := config.DB.Preload("UserParticipants").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": projects})
}

func AddParticipant(c *gin.Context) {
	projectID := c.Param("id")

	var input struct {
		UserID uint `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var project models.Project
	if err := config.DB.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := config.DB.Model(&project).Association("UserParticipants").Append(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config.DB.Preload("UserParticipants").First(&project, project.ID)
	c.JSON(http.StatusOK, gin.H{"data": project})
}

func RemoveParticipant(c *gin.Context) {
	var project models.Project

	projectIDStr := c.Param("id")
	userIDStr := c.Param("user_id")

	var projectID, userID uint
	if _, err := fmt.Sscanf(projectIDStr, "%d", &projectID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := config.DB.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if err := config.DB.Model(&project).Association("UserParticipants").Delete(&models.User{ID: userID}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participant removed successfully"})
}
