package models

import "gorm.io/gorm"

type Task struct {
	gorm.Model
	Title     string `json:"title"`
	Status    string `json:"status"` 
	ProjectID uint   `json:"project_id"`
}
