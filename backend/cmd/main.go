package main

import (
	"github.com/gin-gonic/gin"
	"mundoappointment.com/patients"
	"mundoappointment.com/pkg/config"
)

func main() {
	// Create DB client
	config.CreateClient()

	// Start Server
	router := gin.Default()
	patients.InitPatiantsRoutes(router)
	router.Run()
}
