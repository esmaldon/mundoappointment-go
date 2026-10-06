package patients

import (
	"github.com/gin-gonic/gin"
	"mundoappointment.com/pkg/config"
)

func InitPatientsRoutes(e *gin.RouterGroup, db *config.DBClient) {
	store := NewStore(db)
	service := NewService(store)
	registerPatientsRoutes(e, service)
}

func registerPatientsRoutes(e *gin.RouterGroup, service patientService) {
	h := NewHandler(service)

	e.GET("/clinic/:clinicid/patients", h.getPatients)
	e.GET("/clinic/:clinicid/patients/:id", h.getPatient)
	e.POST("/clinic/:clinicid/patients", h.addPatient)
	e.PATCH("/clinic/:clinicid/patients/:id", h.changePatient)
	e.PATCH("/clinic/:clinicid/patients/:id/parent/:parentid", h.changeParent)
	e.DELETE("/clinic/:clinicid/patients/:id", h.removePatient)
}
