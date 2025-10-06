package controllers

import (
	"kanban/config"
	"kanban/models"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
)

func CreateSprint(c *gin.Context) {
	var input struct {
		Name      string `json:"name"`
		ProjectID uint   `json:"project_id"`
		Goal    string `json:"goal"`
		EstimationType string `json:"estimation_type"`
		TotalEstimation float64 `json:"total_estimation"`
		RemainingEstimation float64 `json:"remaining_estimation"`
		StartDate time.Time `json:"start_date"`
		EndDate time.Time `json:"end_date"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name == "" || input.ProjectID == 0 || input.EstimationType == "" || input.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	sprint := models.Sprint{
		Name:           input.Name,
		ProjectID:      input.ProjectID,
		Goal:          input.Goal,
		EstimationType: input.EstimationType,
		TotalEstimation: input.TotalEstimation,
		RemainingEstimation: input.RemainingEstimation,
		StartDate:      input.StartDate,
		EndDate:        input.EndDate,
		Status:        input.Status,
	}
	if err := config.DB.Create(&sprint).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sprint})
}


func GetAllSprints(c *gin.Context) {
	var sprints []models.Sprint
	if err := config.DB.Preload("Tasks").Find(&sprints).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Update estimasi untuk setiap sprint
	for i := range sprints {
		sprints[i].TotalEstimation = sprints[i].CalculateTotalEstimation()
		sprints[i].RemainingEstimation = sprints[i].CalculateRemainingEstimation()
	}
	
	c.JSON(http.StatusOK, gin.H{"data": sprints})
}

func GetSprint(c *gin.Context) {
	id := c.Param("id")

	var sprint models.Sprint
	if err := config.DB.Where("id = ?", id).Preload("Tasks").First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}
	
	// Update estimasi sprint
	sprint.TotalEstimation = sprint.CalculateTotalEstimation()
	sprint.RemainingEstimation = sprint.CalculateRemainingEstimation()
	
	c.JSON(http.StatusOK, gin.H{"data": sprint})
}

// GetSprintsByProject mendapatkan semua sprint dalam sebuah project
func GetSprintsByProject(c *gin.Context) {
	projectID := c.Param("project_id")

	var sprints []models.Sprint
	if err := config.DB.Where("project_id = ?", projectID).Preload("Tasks").Find(&sprints).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Update estimasi untuk setiap sprint
	for i := range sprints {
		sprints[i].TotalEstimation = sprints[i].CalculateTotalEstimation()
		sprints[i].RemainingEstimation = sprints[i].CalculateRemainingEstimation()
	}
	
	c.JSON(http.StatusOK, gin.H{"data": sprints})
}

// GetSprintAnalytics mendapatkan data analytics untuk chart dan detail sprint
func GetSprintAnalytics(c *gin.Context) {
	id := c.Param("id")

	var sprint models.Sprint
	if err := config.DB.Where("id = ?", id).Preload("Tasks").First(&sprint).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}
	
	// Hitung berbagai metrik
	totalEstimation := sprint.CalculateTotalEstimation()
	remainingEstimation := sprint.CalculateRemainingEstimation()
	completedEstimation := sprint.CalculateCompletedEstimation()
	progressPercentage := sprint.GetProgressPercentage()
	taskBreakdown := sprint.GetTaskStatusBreakdown()
	
	// Data untuk burndown chart (simulasi - dalam implementasi nyata bisa dari historical data)
	burndownData := []map[string]interface{}{
		{"day": 1, "remaining": totalEstimation},
		{"day": 2, "remaining": totalEstimation * 0.9},
		{"day": 3, "remaining": totalEstimation * 0.8},
		{"day": 4, "remaining": remainingEstimation}, // Current day
	}
	
	analytics := map[string]interface{}{
		"sprint_info": map[string]interface{}{
			"id":               sprint.ID,
			"name":             sprint.Name,
			"goal":             sprint.Goal,
			"estimation_type":  sprint.EstimationType,
			"start_date":       sprint.StartDate,
			"end_date":         sprint.EndDate,
			"status":           sprint.Status,
		},
		"estimation_summary": map[string]interface{}{
			"total_estimation":     totalEstimation,
			"remaining_estimation": remainingEstimation,
			"completed_estimation": completedEstimation,
			"progress_percentage":  progressPercentage,
		},
		"task_breakdown": taskBreakdown,
		"burndown_chart": burndownData,
		"tasks": sprint.Tasks,
	}
	
	c.JSON(http.StatusOK, gin.H{"data": analytics})
}

// UpdateSprintStatus mengupdate status sprint
func UpdateSprintStatus(c *gin.Context) {
	id := c.Param("id")
	
	var input struct {
		Status string `json:"status" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var sprint models.Sprint
	if err := config.DB.First(&sprint, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}
	
	sprint.Status = input.Status
	if err := config.DB.Save(&sprint).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"data": sprint})
}