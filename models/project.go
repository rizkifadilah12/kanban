package models

import "gorm.io/gorm"

type Project struct {
	gorm.Model
	Name  string `json:"name"`
	Tasks []Task `json:"tasks"`
}
