package models

import "time"

type Sprint struct {
	ID                  uint      `json:"id" gorm:"primaryKey"`
	ProjectID           uint      `json:"project_id"`
	Name                string    `json:"name"`
	Goal                string    `json:"goal"`
	EstimationType      string    `json:"estimation_type"` // "hour" atau "story_point"
	TotalEstimation     float64   `json:"total_estimation"`
	RemainingEstimation float64   `json:"remaining_estimation"`
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
	Status              string    `json:"status"` // planned, active, completed
	Project             Project   `json:"project" gorm:"foreignKey:ProjectID"`
	Tasks               []Task    `json:"tasks" gorm:"foreignKey:SprintID"`
}

// CalculateTotalEstimation menghitung total estimasi dari semua tasks dalam sprint
func (s *Sprint) CalculateTotalEstimation() float64 {
	var total float64
	for _, task := range s.Tasks {
		total += task.Estimation
	}
	return total
}

// CalculateRemainingEstimation menghitung estimasi yang tersisa berdasarkan task yang belum selesai
func (s *Sprint) CalculateRemainingEstimation() float64 {
	var remaining float64
	for _, task := range s.Tasks {
		// Hanya hitung task yang belum selesai (status bukan "done")
		if task.Status != "done" {
			remaining += task.Estimation
		}
	}
	return remaining
}

// CalculateCompletedEstimation menghitung estimasi yang sudah selesai
func (s *Sprint) CalculateCompletedEstimation() float64 {
	var completed float64
	for _, task := range s.Tasks {
		if task.Status == "done" {
			completed += task.Estimation
		}
	}
	return completed
}

// GetProgressPercentage menghitung persentase progress sprint
func (s *Sprint) GetProgressPercentage() float64 {
	total := s.CalculateTotalEstimation()
	if total == 0 {
		return 0
	}
	completed := s.CalculateCompletedEstimation()
	return (completed / total) * 100
}

// GetTaskStatusBreakdown menghitung breakdown task berdasarkan status
func (s *Sprint) GetTaskStatusBreakdown() map[string]int {
	breakdown := map[string]int{
		"todo":        0,
		"in_progress": 0,
		"done":        0,
	}
	
	for _, task := range s.Tasks {
		breakdown[task.Status]++
	}
	
	return breakdown
}
