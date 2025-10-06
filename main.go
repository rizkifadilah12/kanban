package main

import (
	"kanban/config"
	"kanban/routes"

	"github.com/gin-gonic/gin"

	_ "kanban/docs"
)

// @title Kanban API
// @version 1.0
// @description API untuk aplikasi Kanban Board
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	r := gin.Default()

	config.ConnectDB()

	routes.RegisterRoutes(r)

	r.Run(":8080")
}