package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"mundoappointment.com/clinics"
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
	v1 := router.Group("/api/v1")
	patients.InitPatientsRoutes(v1, db)
	clinics.InitClinicsRoutes(v1, db)
	router.Run()
}
