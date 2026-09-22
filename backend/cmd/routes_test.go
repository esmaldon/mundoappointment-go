package main

import (
	"testing"

	"github.com/gin-gonic/gin"
	"mundoappointment.com/clinics"
	"mundoappointment.com/patients"
)

func TestClinicAndPatientRoutesRegisterTogether(t *testing.T) {
	router := gin.New()
	v1 := router.Group("/api/v1")
	// Registration must not access the database or panic over conflicting parameters.
	patients.InitPatiantsRoutes(v1, nil)
	clinics.InitClinicRoutes(v1, nil)

	want := map[string]bool{
		"GET /api/v1/clinics":                          false,
		"POST /api/v1/clinics/":                        false,
		"GET /api/v1/clinic/:clinicid":                 false,
		"PATCH /api/v1/clinic/:clinicid":               false,
		"DELETE /api/v1/clinic/:clinicid":              false,
		"GET /api/v1/clinic/:clinicid/patients":        false,
		"POST /api/v1/clinic/:clinicid/patients":       false,
		"GET /api/v1/clinic/:clinicid/patients/:id":    false,
		"PATCH /api/v1/clinic/:clinicid/patients/:id":  false,
		"DELETE /api/v1/clinic/:clinicid/patients/:id": false,
	}
	for _, route := range router.Routes() {
		key := route.Method + " " + route.Path
		if _, exists := want[key]; exists {
			want[key] = true
		}
	}
	for route, registered := range want {
		if !registered {
			t.Errorf("missing route: %s", route)
		}
	}
}
