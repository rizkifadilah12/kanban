package models

import "gorm.io/gorm"

type Task struct {
	gorm.Model
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	SprintID    uint   `json:"sprint_id"`
	AssignTo    *uint   `json:"assign_to"`
	Estimation  float64 `json:"estimation"`
	Sprint      Sprint  `json:"sprint"`
	User        User    `json:"user" gorm:"foreignKey:AssignTo"`
}
