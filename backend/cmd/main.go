package main

import (
	"github.com/gin-gonic/gin"
	"mundoappointment.com/patients"
)

func main() {
	router := gin.Default()
	patients.InitPatiantsRoutes(router)
	router.Run()
}
