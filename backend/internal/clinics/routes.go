package clinics

import (
	"github.com/gin-gonic/gin"
	"mundoappointment.com/pkg/config"
)

func InitClinicRoutes(e *gin.RouterGroup, db *config.DBClient) {
	registerClinicRoutes(e, NewStore(db))
}

func registerClinicRoutes(e *gin.RouterGroup, s clinicStore) {
	h := NewHandler(s)

	e.GET("/clinics", h.getClinics)
	e.GET("/clinic/:clinicid", h.getClinic)
	e.POST("/clinics/", h.addClinic)
	e.PATCH("/clinic/:clinicid", h.changeClinic)
	e.DELETE("/clinic/:clinicid", h.removeClinic)
}
