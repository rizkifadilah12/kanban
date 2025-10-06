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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTaskTestDB() {
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

	db.AutoMigrate(&models.User{}, &models.Project{}, &models.Task{})

	config.DB = db
}

func teardownTaskTestDB() {
	if config.DB != nil {
		config.DB.Exec("SET FOREIGN_KEY_CHECKS = 0")
		config.DB.Exec("TRUNCATE TABLE tasks")
		config.DB.Exec("TRUNCATE TABLE project_users")
		config.DB.Exec("TRUNCATE TABLE projects")
		config.DB.Exec("TRUNCATE TABLE users")
		config.DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

		sqlDB, _ := config.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func TestCreateTask(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	project := models.Project{
		Name:        "Test Project",
		Description: "Test",
	}
	config.DB.Create(&project)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	config.DB.Create(&user)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/tasks", CreateTask)
	sprint := models.Sprint{
		ProjectID:      project.ID,
		Name:           "Sprint 1",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
	}
	config.DB.Create(&sprint)

	taskData := map[string]interface{}{
		"title":      "Test Task",
		"status":     "todo",
		"sprint_id":  sprint.ID,
		"assign_to":  &user.ID,
		"estimation": 3.5,
	}

	jsonData, _ := json.Marshal(taskData)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var task models.Task
	err := config.DB.Where("title = ?", "Test Task").First(&task).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Task", task.Title)
	assert.Equal(t, "todo", task.Status)
	assert.Equal(t, sprint.ID, task.SprintID)
	assert.Equal(t, user.ID, *task.AssignTo)
}

func TestCreateTaskWithoutAssignee(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	project := models.Project{
		Name: "Test Project",
	}
	config.DB.Create(&project)
	sprint := models.Sprint{
		ProjectID:      project.ID,
		Name:           "Sprint 1",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
	}
	config.DB.Create(&sprint)
	// fmt.Println("Created sprint with ID:", sprint.ID)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/tasks", CreateTask)

	taskData := map[string]interface{}{
		"title":      "Unassigned Task",
		"status":     "todo",
		"sprint_id":  sprint.ID,
		"estimation": 2.0,
	}

	jsonData, _ := json.Marshal(taskData)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var task models.Task
	config.DB.Where("title = ?", "Unassigned Task").First(&task)
	assert.Nil(t, task.AssignTo)
}

func TestCreateTaskInvalidJSON(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/tasks", CreateTask)

	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestCreateTaskMissingRequiredFields(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/tasks", CreateTask)

	taskData := map[string]interface{}{
		"status":     "todo",
		"sprint_id":  1,
	}

	jsonData, _ := json.Marshal(taskData)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Missing required fields")
}

func TestGetAllTasks(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	project := models.Project{Name: "Test Project"}
	config.DB.Create(&project)

	sprint := models.Sprint{
		ProjectID:      project.ID,
		Name:           "Sprint 1",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
	}
	config.DB.Create(&sprint)

	task1 := models.Task{
		Title:      "Task 1",
		Status:     "todo",
		SprintID:   sprint.ID,
		Estimation: 2.0,
	}
	task2 := models.Task{
		Title:      "Task 2",
		Status:     "in_progress",
		SprintID:   sprint.ID,
		Estimation: 4.0,
	}
	config.DB.Create(&task1)
	config.DB.Create(&task2)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tasks", GetAllTasks)

	req, _ := http.NewRequest("GET", "/tasks", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string][]models.Task
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.GreaterOrEqual(t, len(response["data"]), 2)
}

func TestGetAllTasksEmpty(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tasks", GetAllTasks)

	req, _ := http.NewRequest("GET", "/tasks", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string][]models.Task
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, 0, len(response["data"]))
}

func TestGetTasks(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	project1 := models.Project{Name: "Project 1"}
	project2 := models.Project{Name: "Project 2"}
	config.DB.Create(&project1)
	config.DB.Create(&project2)

	sprint1 := models.Sprint{
		ProjectID:      project1.ID,
		Name:           "Sprint 1",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
	}
	sprint2 := models.Sprint{
		ProjectID:      project2.ID,
		Name:           "Sprint 1",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
	}
	config.DB.Create(&sprint1)
	config.DB.Create(&sprint2)
	task1 := models.Task{Title: "Task 1", Status: "todo",  SprintID: sprint1.ID, Estimation: 3.0}
	task2 := models.Task{Title: "Task 2", Status: "todo", SprintID: sprint1.ID, Estimation: 5.0}
	task3 := models.Task{Title: "Task 3", Status: "todo", SprintID: sprint2.ID, Estimation: 2.0}
	config.DB.Create(&task1)
	config.DB.Create(&task2)
	config.DB.Create(&task3)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/projects/:id/tasks", GetTasks)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/projects/%d/tasks", sprint1.ID), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string][]models.Task
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response["data"]))
}

func TestUpdateTaskStatus(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	project := models.Project{Name: "Test Project"}
	config.DB.Create(&project)
	sprint := models.Sprint{
		ProjectID:      project.ID,
		Name:           "Sprint 1",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
	}
	config.DB.Create(&sprint)
	task := models.Task{
		Title:      "Test Task",
		Status:     "todo",
		SprintID:   sprint.ID,
		Estimation: 5.0,
	}
	config.DB.Create(&task)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/tasks/:id", UpdateTaskStatus)

	updateData := map[string]string{
		"status": "in_progress",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/tasks/%d", task.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var updatedTask models.Task
	config.DB.First(&updatedTask, task.ID)
	assert.Equal(t, "in_progress", updatedTask.Status)
}

func TestUpdateTaskStatusNotFound(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/tasks/:id", UpdateTaskStatus)

	updateData := map[string]string{
		"status": "done",
	}

	jsonData, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/tasks/99999", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestUpdateTaskStatusInvalidJSON(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	project := models.Project{Name: "Test"}
	config.DB.Create(&project)
	sprint := models.Sprint{
		ProjectID:      project.ID,
		Name:           "Sprint 1",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
	}
	config.DB.Create(&sprint)
	task := models.Task{Title: "Test", Status: "todo", SprintID: sprint.ID, Estimation: 5.0}
	config.DB.Create(&task)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/tasks/:id", UpdateTaskStatus)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/tasks/%d", task.ID), bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestAssignToUser(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Username: "assignee",
		Password: string(hashedPassword),
	}
	config.DB.Create(&user)

	project := models.Project{Name: "Test Project"}
	config.DB.Create(&project)
	sprint := models.Sprint{
		ProjectID:      project.ID,
		Name:           "Sprint 1",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
	}
	config.DB.Create(&sprint)
	task := models.Task{
		Title:      "Unassigned Task",
		Status:     "todo",
		SprintID:   sprint.ID,
		Estimation: 5.0,
	}
	config.DB.Create(&task)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/tasks/:id/assign", AssignToUser)

	assignData := map[string]uint{
		"assign_to": user.ID,
	}

	jsonData, _ := json.Marshal(assignData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/tasks/%d/assign", task.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var updatedTask models.Task
	config.DB.First(&updatedTask, task.ID)
	assert.Equal(t, user.ID, *updatedTask.AssignTo)
}

func TestAssignToUserNotFound(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/tasks/:id/assign", AssignToUser)

	assignData := map[string]uint{
		"assign_to": 1,
	}

	jsonData, _ := json.Marshal(assignData)
	req, _ := http.NewRequest("PUT", "/tasks/99999/assign", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestAssignToUserUnassign(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Username: "user1",
		Password: string(hashedPassword),
	}
	config.DB.Create(&user)

	project := models.Project{Name: "Test"}
	config.DB.Create(&project)
	sprint := models.Sprint{
		ProjectID:      project.ID,
		Name:           "Sprint 1",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
	}
	config.DB.Create(&sprint)
	task := models.Task{
		Title:     "Assigned Task",
		Status:    "todo",
		SprintID:  sprint.ID,
		AssignTo:  &user.ID,
	}
	config.DB.Create(&task)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/tasks/:id/assign", AssignToUser)


	assignData := map[string]uint{
		"assign_to": 0,
	}

	jsonData, _ := json.Marshal(assignData)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/tasks/%d/assign", task.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var updatedTask models.Task
	config.DB.First(&updatedTask, task.ID)
	assert.Nil(t, updatedTask.AssignTo)
}

func TestDeleteTask(t *testing.T) {
	setupTaskTestDB()
	defer teardownTaskTestDB()

	project := models.Project{Name: "Test"}
	config.DB.Create(&project)
	sprint := models.Sprint{
		ProjectID:      project.ID,
		Name:           "Sprint 1",
		EstimationType: "hour",
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(0, 0, 7),
	}
	config.DB.Create(&sprint)
	task := models.Task{Title: "Test", Status: "todo", SprintID: sprint.ID, Estimation: 5.0}
	config.DB.Create(&task)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/tasks/:id", DeleteTask)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/tasks/%d", task.ID), nil)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var deletedTask models.Task
	err := config.DB.First(&deletedTask, task.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}
