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

func setupTestDB() {
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

func teardownTestDB() {
	if config.DB != nil {
		config.DB.Exec("SET FOREIGN_KEY_CHECKS = 0")
		config.DB.Exec("TRUNCATE TABLE tasks")
		config.DB.Exec("TRUNCATE TABLE projects")
		config.DB.Exec("TRUNCATE TABLE users")
		config.DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

		sqlDB, _ := config.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

func TestRegister(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", Register)

	user := models.User{
		Username: "testuser",
		Password: "testpass123",
	}

	jsonData, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var response map[string]string
	json.Unmarshal(resp.Body.Bytes(), &response)
	assert.Equal(t, "registered", response["message"])

	var dbUser models.User
	err := config.DB.Where("username = ?", "testuser").First(&dbUser).Error
	assert.NoError(t, err)
	assert.Equal(t, "testuser", dbUser.Username)
	assert.NotEqual(t, "testpass123", dbUser.Password)
}

func TestRegisterDuplicateUsername(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.DefaultCost)
	existingUser := models.User{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	config.DB.Create(&existingUser)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", Register)

	user := models.User{
		Username: "testuser",
		Password: "anotherpass",
	}

	jsonData, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.NotEqual(t, http.StatusOK, resp.Code)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, resp.Code)
}

func TestRegisterInvalidJSON(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/register", Register)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestLogin(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{
		Username: "logintest",
		Password: string(hashedPassword),
	}
	config.DB.Create(&user)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", Login)

	loginData := models.User{
		Username: "logintest",
		Password: password,
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestLoginInvalidCredentials(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", Login)

	loginData := models.User{
		Username: "nonexistent",
		Password: "wrongpass",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestLoginWrongPassword(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	user := models.User{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	config.DB.Create(&user)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", Login)

	loginData := models.User{
		Username: "testuser",
		Password: "wrongpassword",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestLoginInvalidJSON(t *testing.T) {
	setupTestDB()
	defer teardownTestDB()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", Login)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}
