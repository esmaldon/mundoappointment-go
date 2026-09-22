package patients

import (
	"github.com/gin-gonic/gin"
	"mundoappointment.com/pkg/config"
)

func InitPatiantsRoutes(e *gin.RouterGroup, db *config.DBClient) {
	s := NewStore(db)
	h := NewHandler(s)

	e.GET("/clinic/:clinicid/patients", h.getPatients)
	e.GET("/clinic/:clinicid/patients/:id", h.getPatient)
	e.POST("/clinic/:clinicid/patients", h.addPatient)
	e.PATCH("/clinic/:clinicid/patients/:id", h.changePatient)
	e.DELETE("/clinic/:clinicid/patients/:id", h.removePatient)
}
