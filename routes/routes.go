package routes

import (
	"kanban/controllers"
	"kanban/middlewares"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(r *gin.Engine) {
	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Authentication
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	auth := r.Group("/")
	auth.Use(middlewares.AuthMiddleware())
	{
		// Project
		auth.POST("/projects", controllers.CreateProject)
		auth.GET("/projects", controllers.GetAllProjects)
		auth.GET("/projects/:id", controllers.GetProjects)
		auth.POST("/projects/:id/participants", controllers.AddParticipant)
		auth.DELETE("/projects/:id/participants/:user_id", controllers.RemoveParticipant)

		// Task
		auth.POST("/tasks", controllers.CreateTask)
		auth.PUT("/tasks/:id", controllers.UpdateTaskStatus)
		auth.GET("/tasks", controllers.GetAllTasks)
		auth.GET("/tasks/:id", controllers.GetTasks)
		auth.PUT("/tasks/:id/assign", controllers.AssignToUser)
		auth.DELETE("/tasks/:id", controllers.DeleteTask)

		// Sprint
		auth.POST("/sprints", controllers.CreateSprint)
		auth.GET("/sprints", controllers.GetAllSprints)
		auth.GET("/sprints/:id", controllers.GetSprint)
		auth.GET("/sprints/:id/analytics", controllers.GetSprintAnalytics)
		auth.PUT("/sprints/:id/status", controllers.UpdateSprintStatus)
		auth.GET("/projects/:project_id/sprints", controllers.GetSprintsByProject)
	}

}
