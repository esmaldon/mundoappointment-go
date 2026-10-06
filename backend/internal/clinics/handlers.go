package clinics

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"mundoappointment.com/pkg/httpmessage"
)

type clinicService interface {
	getClinics() ([]Clinic, error)
	getClinic(string) (Clinic, error)
	addClinic(CreateClinicRequest) (Clinic, error)
	changeClinic(UpdateClinicRequest, string) (Clinic, error)
	removeClinic(string) (string, error)
}

type handler struct {
	s clinicService
}

func NewHandler(s clinicService) *handler {
	return &handler{
		s: s,
	}
}

func (h *handler) getClinics(c *gin.Context) {
	clinics, err := h.s.getClinics()
	if err != nil {
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusOK, clinics)
}

func (h *handler) getClinic(c *gin.Context) {
	id := c.Param("clinicid")
	clinic, err := h.s.getClinic(id)
	if err != nil {
		if errors.Is(err, ErrorClinicNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusOK, clinic)
}

func (h *handler) addClinic(c *gin.Context) {
	var req CreateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpmessage.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	clinic, err := h.s.addClinic(req)
	if err != nil {
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusCreated, clinic)
}

func (h *handler) changeClinic(c *gin.Context) {
	id := c.Param("clinicid")
	var req UpdateClinicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpmessage.Fail(c, http.StatusBadRequest, "BAD_REQUEST", err.Error())
		return
	}
	if req.IsEmpty() {
		httpmessage.Fail(c, http.StatusBadRequest, "EMPTY_REQUEST", "at least one field should be provided for update")
		return
	}
	clinicUpdated, err := h.s.changeClinic(req, id)
	if err != nil {
		if errors.Is(err, ErrorClinicNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusOK, clinicUpdated)
}

func (h *handler) removeClinic(c *gin.Context) {
	id := c.Param("clinicid")
	clinicId, err := h.s.removeClinic(id)
	if err != nil {
		if errors.Is(err, ErrorClinicNotFound) {
			httpmessage.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		httpmessage.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	httpmessage.Success(c, http.StatusOK, clinicId)
}
