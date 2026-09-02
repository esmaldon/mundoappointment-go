package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"mundoappointment.com/patients"
	"mundoappointment.com/pkg/config"
)

func main() {
	// Create DB client
	db, err := config.NewDBClient()
	if err != nil {
		log.Fatalf("DB Connection failure %v", err)
	}
	// Start Server
	router := gin.Default()
	patients.InitPatiantsRoutes(router, db)
	router.Run()
}
