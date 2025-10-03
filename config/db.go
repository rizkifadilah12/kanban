package config

import (
	"log"
	"kanban/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)
var DB *gorm.DB

func ConnectDB() {
	dsn := "root:admin123@tcp(127.0.0.1:3306)/gin_api?charset=utf8mb4&parseTime=True&loc=Local"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Gagal konek database:", err)
	}

	// migrasi tabel
	database.AutoMigrate(&models.Project{}, &models.Task{})

	DB = database
}