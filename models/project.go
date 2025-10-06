package models

import "gorm.io/gorm"

type Project struct {
	gorm.Model
	Name             string `json:"name"`
	Description      string `json:"description"`
	UserParticipants []User  `gorm:"many2many:project_users;"`
	
}
