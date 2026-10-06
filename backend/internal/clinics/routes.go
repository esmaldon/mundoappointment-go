package clinics

import (
	"github.com/gin-gonic/gin"
	"mundoappointment.com/pkg/config"
)

func InitClinicsRoutes(e *gin.RouterGroup, db *config.DBClient) {
	store := NewStore(db)
	service := NewService(store)
	registerClinicsRoutes(e, service)
}

func registerClinicsRoutes(e *gin.RouterGroup, service clinicService) {
	h := NewHandler(service)

	e.GET("/clinics", h.getClinics)
	e.GET("/clinic/:clinicid", h.getClinic)
	e.POST("/clinics/", h.addClinic)
	e.PATCH("/clinic/:clinicid", h.changeClinic)
	e.DELETE("/clinic/:clinicid", h.removeClinic)
}
