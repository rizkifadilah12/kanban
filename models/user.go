package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID       uint      `json:"id"`
	Username string    `json:"username" gorm:"unique"`
	Password string    `json:"password"`
	Email    string    `json:"email" gorm:"unique"`
	Projects []Project `json:"projects,omitempty" gorm:"many2many:project_users;"`
}
