package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"kanban/config"
	"kanban/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupSprintTestDB() {
	godotenv.Load("../.env")

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	testDBName := os.Getenv("TEST_DB_NAME")
	if testDBName == "" {
		testDBName = "kanban_test"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, testDBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect test database:", err)
	}

	db.AutoMigrate(&models.User{}, &models.Project{}, &models.Sprint{}, &models.Task{})

	config.DB = db
}

func teardownSprintTestDB() {
	if config.DB != nil {
		config.DB.Exec("SET FOREIGN_KEY_CHECKS = 0")
		config.DB.Exec("TRUNCATE TABLE tasks")
		config.DB.Exec("TRUNCATE TABLE sprints")
		config.DB.Exec("TRUNCATE TABLE projects")
		config.DB.Exec("TRUNCATE TABLE users")
		config.DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

		sqlDB, _ := config.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func TestCreateSprintWithHourEstimation(t *testing.T) {
	setupSprintTestDB()
	defer teardownSprintTestDB()

	project := models.Project{
		Name: "Test Project",
	}
	config.DB.Create(&project)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/sprints", CreateSprint)

	sprintData := map[string]interface{}{
		"name":                 "Sprint 1",
		"project_id":           project.ID,
		"goal":                 "Complete user authentication",
		"estimation_type":      "hour",
		"total_estimation":     40.0,
		"remaining_estimation": 40.0,
		"start_date":           time.Now().Format(time.RFC3339),
		"end_date":             time.Now().AddDate(0, 0, 14).Format(time.RFC3339),
		"status":               "planned",
	}

	jsonData, _ := json.Marshal(sprintData)
	req, _ := http.NewRequest("POST", "/sprints", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]models.Sprint
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Sprint 1", response["data"].Name)
	assert.Equal(t, "hour", response["data"].EstimationType)
	assert.Equal(t, 40.0, response["data"].TotalEstimation)
}

func TestCreateSprintWithStoryPointEstimation(t *testing.T) {
	setupSprintTestDB()
	defer teardownSprintTestDB()

	project := models.Project{
		Name: "Test Project",
	}
	config.DB.Create(&project)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/sprints", CreateSprint)

	sprintData := map[string]interface{}{
		"name":                 "Sprint 2",
		"project_id":           project.ID,
		"goal":                 "Complete user dashboard",
		"estimation_type":      "story_point",
		"total_estimation":     21.0,
		"remaining_estimation": 21.0,
		"start_date":           time.Now().Format(time.RFC3339),
		"end_date":             time.Now().AddDate(0, 0, 14).Format(time.RFC3339),
		"status":               "planned",
	}

	jsonData, _ := json.Marshal(sprintData)
	req, _ := http.NewRequest("POST", "/sprints", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]models.Sprint
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Sprint 2", response["data"].Name)
	assert.Equal(t, "story_point", response["data"].EstimationType)
	assert.Equal(t, 21.0, response["data"].TotalEstimation)
}

func TestGetSprintWithEstimationCalculation(t *testing.T) {
	setupSprintTestDB()
	defer teardownSprintTestDB()

	project := models.Project{Name: "Test Project"}
	config.DB.Create(&project)

	sprint := models.Sprint{
		Name:           "Test Sprint",
		ProjectID:      project.ID,
		Goal:           "Test Goal",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
		Status:         "active",
	}
	config.DB.Create(&sprint)

	task1 := models.Task{
		Title:      "Task 1",
		Status:     "todo",
		SprintID:   sprint.ID,
		Estimation: 5.0,
	}
	task2 := models.Task{
		Title:      "Task 2",
		Status:     "done",
		SprintID:   sprint.ID,
		Estimation: 3.0,
	}
	task3 := models.Task{
		Title:      "Task 3",
		Status:     "in_progress",
		SprintID:   sprint.ID,
		Estimation: 2.0,
	}
	config.DB.Create(&task1)
	config.DB.Create(&task2)
	config.DB.Create(&task3)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/sprints/:id", GetSprint)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/sprints/%d", sprint.ID), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]models.Sprint
	json.Unmarshal(resp.Body.Bytes(), &response)

	sprintData := response["data"]
	assert.Equal(t, 10.0, sprintData.TotalEstimation)
	assert.Equal(t, 7.0, sprintData.RemainingEstimation)
}

func TestGetSprintAnalytics(t *testing.T) {
	setupSprintTestDB()
	defer teardownSprintTestDB()

	project := models.Project{Name: "Test Project"}
	config.DB.Create(&project)

	sprint := models.Sprint{
		Name:           "Analytics Sprint",
		ProjectID:      project.ID,
		Goal:           "Test Analytics",
		EstimationType: "story_point",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
		Status:         "active",
	}
	config.DB.Create(&sprint)

	task1 := models.Task{Title: "Task 1", Status: "todo", SprintID: sprint.ID, Estimation: 8.0}
	task2 := models.Task{Title: "Task 2", Status: "done", SprintID: sprint.ID, Estimation: 5.0}
	task3 := models.Task{Title: "Task 3", Status: "in_progress", SprintID: sprint.ID, Estimation: 3.0}
	task4 := models.Task{Title: "Task 4", Status: "done", SprintID: sprint.ID, Estimation: 2.0}
	config.DB.Create(&task1)
	config.DB.Create(&task2)
	config.DB.Create(&task3)
	config.DB.Create(&task4)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/sprints/:id/analytics", GetSprintAnalytics)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/sprints/%d/analytics", sprint.ID), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]interface{}
	json.Unmarshal(resp.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})

	sprintInfo := data["sprint_info"].(map[string]interface{})
	assert.Equal(t, "Analytics Sprint", sprintInfo["name"])
	assert.Equal(t, "story_point", sprintInfo["estimation_type"])

	estimationSummary := data["estimation_summary"].(map[string]interface{})
	assert.Equal(t, 18.0, estimationSummary["total_estimation"])
	assert.Equal(t, 11.0, estimationSummary["remaining_estimation"])
	assert.Equal(t, 7.0, estimationSummary["completed_estimation"])

	taskBreakdown := data["task_breakdown"].(map[string]interface{})
	assert.Equal(t, float64(1), taskBreakdown["todo"])
	assert.Equal(t, float64(1), taskBreakdown["in_progress"])
	assert.Equal(t, float64(2), taskBreakdown["done"])

	assert.NotNil(t, data["burndown_chart"])
}

func TestUpdateSprintStatus(t *testing.T) {
	setupSprintTestDB()
	defer teardownSprintTestDB()

	project := models.Project{Name: "Test Project"}
	config.DB.Create(&project)

	sprint := models.Sprint{
		Name:           "Status Test Sprint",
		ProjectID:      project.ID,
		Goal:           "Test Status Update",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
		Status:         "planned",
	}
	config.DB.Create(&sprint)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/sprints/:id/status", UpdateSprintStatus)

	updateData := map[string]string{
		"status": "active",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/sprints/%d/status", sprint.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var updatedSprint models.Sprint
	config.DB.First(&updatedSprint, sprint.ID)
	assert.Equal(t, "active", updatedSprint.Status)
}

func TestGetSprintsByProject(t *testing.T) {
	setupSprintTestDB()
	defer teardownSprintTestDB()

	project1 := models.Project{Name: "Project 1"}
	project2 := models.Project{Name: "Project 2"}
	config.DB.Create(&project1)
	config.DB.Create(&project2)

	sprint1 := models.Sprint{Name: "Sprint 1", ProjectID: project1.ID, EstimationType: "hour", Status: "active", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 7)}
	sprint2 := models.Sprint{Name: "Sprint 2", ProjectID: project1.ID, EstimationType: "story_point", Status: "planned", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14)}
	sprint3 := models.Sprint{Name: "Sprint 3", ProjectID: project2.ID, EstimationType: "hour", Status: "completed", StartDate: time.Now().AddDate(0, 0, -14), EndDate: time.Now().AddDate(0, 0, -7)}
	config.DB.Create(&sprint1)
	config.DB.Create(&sprint2)
	config.DB.Create(&sprint3)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/projects/:project_id/sprints", GetSprintsByProject)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/projects/%d/sprints", project1.ID), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string][]models.Sprint
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response["data"]))
}

func TestCreateSprintMissingRequiredFields(t *testing.T) {
	setupSprintTestDB()
	defer teardownSprintTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/sprints", CreateSprint)

	sprintData := map[string]interface{}{
		"name": "Incomplete Sprint",
	}

	jsonData, _ := json.Marshal(sprintData)
	req, _ := http.NewRequest("POST", "/sprints", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Missing required fields")
}

func TestGetSprintNotFound(t *testing.T) {
	setupSprintTestDB()
	defer teardownSprintTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/sprints/:id", GetSprint)

	req, _ := http.NewRequest("GET", "/sprints/99999", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
}
