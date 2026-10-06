package patients

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"mundoappointment.com/pkg/httpmessage"
)

type patientService interface {
	getPatients(string) ([]Patient, error)
	getPatient(string, string) (Patient, error)
	addPatient(CreatePatientRequest, string) (Patient, error)
	addParent(CreatePatientRequest) (Parent, error)
	changePatient(UpdatePatientRequest, string, string) (Patient, error)
	changeParent(UpdateParentRequest, string, string, string) (Parent, error)
	removePatient(string, string) (string, error)
	removeParent(string) (string, error)
}

type handler struct {
	s patientService
}

func NewHandler(s patientService) *handler {
	return &handler{
		s: s,
	}
}

func (h *handler) getPatients(c *gin.Context) {
	clinicId := c.Param("clinicid")
	patients, err := h.s.getPatients(clinicId)
	if err != nil {
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusOK, patients)
}

func (h *handler) getPatient(c *gin.Context) {
	id := c.Param("id")
	clinicId := c.Param("clinicid")
	patient, err := h.s.getPatient(id, clinicId)
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
	clinicId := c.Param("clinicid")
	var req CreatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpmessage.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	patient, err := h.s.addPatient(req, clinicId)
	if err != nil {
		if errors.Is(err, ErrorPatientBadReq) || errors.Is(err, ErrorParentEmpty) {
			httpmessage.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusCreated, patient)
}

func (h *handler) changePatient(c *gin.Context) {
	id := c.Param("id")
	clinicId := c.Param("clinicid")
	var req UpdatePatientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpmessage.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if req.IsEmpty() {
		httpmessage.Fail(c, http.StatusBadRequest, "EMPTY_REQUEST", "at least one field should be provided for update")
		return
	}
	patientUpdated, err := h.s.changePatient(req, id, clinicId)
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusOK, patientUpdated)
}

func (h *handler) changeParent(c *gin.Context) {
	id := c.Param("parentid")
	var req UpdateParentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpmessage.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if req.IsEmpty() {
		httpmessage.Fail(c, http.StatusBadRequest, "EMPTY_REQUEST", "at least one field should be provided for update")
		return
	}
	parentUpdated, err := h.s.changeParent(req, id, c.Param("id"), c.Param("clinicid"))
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) || errors.Is(err, ErrorParentNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusOK, parentUpdated)
}

func (h *handler) removePatient(c *gin.Context) {
	id := c.Param("id")
	clinicId := c.Param("clinicid")
	patientId, err := h.s.removePatient(id, clinicId)
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
