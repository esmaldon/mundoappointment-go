package patients

import (
	"github.com/gin-gonic/gin"
	"mundoappointment.com/pkg/config"
)

func InitPatiantsRoutes(e *gin.Engine, db *config.DBClient) {
	s := NewStore(db)
	h := NewHandler(s)

	e.GET("/patients", h.getPatients)
	e.GET("/patients/:id", h.getPatient)
	e.POST("/patients", h.addPatient)
	e.PUT("/patients/:id", h.changePatient)
	e.DELETE("/patients/:id", h.removePatient)
}
