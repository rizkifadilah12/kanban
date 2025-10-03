package main

import (
	"github.com/gin-gonic/gin"
	"kanban/config"
	"kanban/routes"
)


func main() {
	r := gin.Default()

	config.ConnectDB()

	routes.RegisterRoutes(r)

	r.Run(":8080")
}