package patients

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"mundoappointment.com/pkg/httpmessage"
)

type handler struct {
	s *store
}

func NewHandler(s *store) *handler {
	return &handler{
		s: s,
	}
}

func (h *handler) getPatients(c *gin.Context) {
	patients, err := h.s.fetchPatients()
	if err != nil {
		httpmessage.Fail(c, http.StatusInternalServerError, "500", err.Error())
		return
	}
	httpmessage.OK(c, patients)
}

func (h *handler) getPatient(c *gin.Context) {
	id := c.Param("id")
	patient, err := h.s.fetchPatient(id)
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "404", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "500", err.Error())
		return
	}

	httpmessage.OK(c, patient)
}

func (h *handler) addPatient(c *gin.Context) {
	var reqPatient Patient
	if err := c.BindJSON(&reqPatient); err != nil {
		httpmessage.Fail(c, http.StatusInternalServerError, "500", err.Error())
		return
	}
	patient, err := h.s.createPatient(reqPatient)
	if err != nil {
		httpmessage.Fail(c, http.StatusInternalServerError, "500", err.Error())
		return
	}
	httpmessage.OK(c, patient)
}

func (h *handler) removePatient(c *gin.Context) {
	id := c.Param("id")
	patientId, err := h.s.deletePatient(id)
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "404", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "500", err.Error())
		return
	}
	httpmessage.OK(c, patientId)
}

func (h *handler) changePatient(c *gin.Context) {
	id := c.Param("id")
	var reqPatient Patient
	if err := c.BindJSON(&reqPatient); err != nil {
		httpmessage.Fail(c, http.StatusInternalServerError, "500", err.Error())
		return
	}
	patientUpdated, err := h.s.updatePatient(id, reqPatient)
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "404", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "500", err.Error())
		return
	}
	c.JSON(http.StatusOK, patientUpdated)
}
