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
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusOK, patients)
}

func (h *handler) getPatient(c *gin.Context) {
	id := c.Param("id")
	patient, err := h.s.fetchPatient(id)
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	httpmessage.Success(c, http.StatusOK, patient)
}

func (h *handler) addPatient(c *gin.Context) {
	var req CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpmessage.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	patient, err := h.s.createPatient(req)
	if err != nil {
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusCreated, patient)
}

func (h *handler) removePatient(c *gin.Context) {
	id := c.Param("id")
	patientId, err := h.s.deletePatient(id)
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusOK, patientId)
}

func (h *handler) changePatient(c *gin.Context) {
	id := c.Param("id")
	var req UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpmessage.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if req.IsEmpty() {
		httpmessage.Fail(c, http.StatusBadRequest, "EMPTY_REQUEST", "at least one field should be provided for update")
		return
	}
	patientUpdated, err := h.s.updatePatient(id, req)
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	c.JSON(http.StatusOK, patientUpdated)
}
