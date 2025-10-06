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

	"kanban/config"
	"kanban/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupProjectTestDB() {
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

func teardownProjectTestDB() {
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

func TestCreateProject(t *testing.T) {
	setupProjectTestDB()
	defer teardownProjectTestDB()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user1 := models.User{
		Username: "user1",
		Password: string(hashedPassword),
		Email:    "user1@example.com",
	}
	user2 := models.User{
		Username: "user2",
		Password: string(hashedPassword),
		Email:    "user2@example.com",
	}
	config.DB.Create(&user1)
	config.DB.Create(&user2)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/projects", CreateProject)
	projectData := map[string]any{
		"name":            "Test Project",
		"description":     "This is a test project",
		"hours":           10,
		"story_points":    5,
		"participant_ids": []uint{user1.ID, user2.ID},
	}

	jsonData, _ := json.Marshal(projectData)
	req, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)
	var response map[string]any
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NotNil(t, response["data"])

	var project models.Project
	err := config.DB.Preload("UserParticipants").Where("name = ?", "Test Project").First(&project).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, 2, len(project.UserParticipants))
}

func TestCreateProjectWithoutParticipants(t *testing.T) {
	setupProjectTestDB()
	defer teardownProjectTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/projects", CreateProject)

	projectData := map[string]interface{}{
		"name":        "Solo Project",
		"description": "Project without participants",
		"hours":       5,
		"story_points": 3,
	}

	jsonData, _ := json.Marshal(projectData)
	req, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	var project models.Project
	err := config.DB.Preload("UserParticipants").Where("name = ?", "Solo Project").First(&project).Error
	assert.NoError(t, err)
	assert.Equal(t, 0, len(project.UserParticipants))
}

func TestCreateProjectInvalidJSON(t *testing.T) {
	setupProjectTestDB()
	defer teardownProjectTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/projects", CreateProject)

	req, _ := http.NewRequest("POST", "/projects", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestCreateProjectMissingRequiredField(t *testing.T) {
	setupProjectTestDB()
	defer teardownProjectTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/projects", CreateProject)

	projectData := map[string]interface{}{
		"description": "No name provided",
		"hours":       10,
	}

	jsonData, _ := json.Marshal(projectData)
	req, _ := http.NewRequest("POST", "/projects", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetAllProjects(t *testing.T) {
	setupProjectTestDB()
	defer teardownProjectTestDB()

	project1 := models.Project{
		Name:        "Project 1",
		Description: "First project",
	}
	project2 := models.Project{
		Name:        "Project 2",
		Description: "Second project",
	}
	config.DB.Create(&project1)
	config.DB.Create(&project2)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/projects", GetAllProjects)

	req, _ := http.NewRequest("GET", "/projects", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "application/json; charset=utf-8", resp.Header().Get("Content-Type"))

	var response map[string][]models.Project
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.GreaterOrEqual(t, len(response["data"]), 2)
}

func TestGetAllProjectsEmpty(t *testing.T) {
	setupProjectTestDB()
	defer teardownProjectTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/projects", GetAllProjects)

	req, _ := http.NewRequest("GET", "/projects", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string][]models.Project
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, 0, len(response["data"]))
}

func TestGetProject(t *testing.T) {
	setupProjectTestDB()
	defer teardownProjectTestDB()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	config.DB.Create(&user)

	project := models.Project{
		Name:        "Test Project",
		Description: "Get project test",
	}
	config.DB.Create(&project)
	config.DB.Model(&project).Association("UserParticipants").Append(&user)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/projects/:id", GetProjects)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/projects/%d", project.ID), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]models.Project
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "Test Project", response["data"].Name)
	assert.Equal(t, 1, len(response["data"].UserParticipants))
}

func TestGetProjectNotFound(t *testing.T) {
	setupProjectTestDB()
	defer teardownProjectTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/projects/:id", GetProjects)

	req, _ := http.NewRequest("GET", "/projects/99999", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestAddParticipant(t *testing.T) {
	setupProjectTestDB()
	defer teardownProjectTestDB()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Username: "newuser",
		Password: string(hashedPassword),
	}
	config.DB.Create(&user)

	project := models.Project{
		Name: "Test Project",
	}
	config.DB.Create(&project)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/projects/:id/participants", AddParticipant)

	participantData := map[string]uint{
		"user_id": user.ID,
	}

	jsonData, _ := json.Marshal(participantData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/projects/%d/participants", project.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var updatedProject models.Project
	config.DB.Preload("UserParticipants").First(&updatedProject, project.ID)
	assert.Equal(t, 1, len(updatedProject.UserParticipants))
}

func TestRemoveParticipant(t *testing.T) {
	setupProjectTestDB()
	defer teardownProjectTestDB()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Username: "removeuser",
		Password: string(hashedPassword),
	}
	config.DB.Create(&user)

	project := models.Project{
		Name: "Test Project",
	}
	config.DB.Create(&project)
	config.DB.Model(&project).Association("UserParticipants").Append(&user)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/projects/:id/participants/:user_id", RemoveParticipant)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/projects/%d/participants/%d", project.ID, user.ID), nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var updatedProject models.Project
	config.DB.Preload("UserParticipants").First(&updatedProject, project.ID)
	assert.Equal(t, 0, len(updatedProject.UserParticipants))
}
