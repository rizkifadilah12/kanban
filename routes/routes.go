package routes

import (
	"kanban/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	// Project
	r.POST("/projects", controllers.CreateProject)
	r.GET("/projects", controllers.GetProjects)

	// Task
	r.POST("/tasks", controllers.CreateTask)
	r.PUT("/tasks/:id", controllers.UpdateTaskStatus)
}
