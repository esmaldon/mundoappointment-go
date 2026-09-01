package patients

import (
	"github.com/gin-gonic/gin"
)

func InitPatiantsRoutes(e *gin.Engine) {
	e.GET("/patients", getPatients)
	e.GET("/patients/:id", getPatient)
	e.POST("/patients", addPatient)
	e.PUT("/patients", changePatient)
	e.DELETE("/patients/:id", removePatient)
}
