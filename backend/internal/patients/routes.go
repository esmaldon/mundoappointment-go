package patients

import (
	"github.com/gin-gonic/gin"
)

func InitPatiantsRoutes(e *gin.Engine) {
	e.GET("/patients", getPatients)
	e.POST("/patients", addPatient)
}
