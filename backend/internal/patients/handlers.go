package patients

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, patients)
}

func (h *handler) getPatient(c *gin.Context) {
	id := c.Param("id")
	patient, err := h.s.fetchPatient(id)
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, patient)
}

func (h *handler) addPatient(c *gin.Context) {
	var reqPatient Patient
	if err := c.BindJSON(&reqPatient); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	patient, err := h.s.createPatient(reqPatient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, patient)
}

func (h *handler) removePatient(c *gin.Context) {
	id := c.Param("id")
	patientId, err := h.s.deletePatient(id)
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, patientId)
}

func (h *handler) changePatient(c *gin.Context) {
	id := c.Param("id")
	var reqPatient Patient
	if err := c.BindJSON(&reqPatient); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	patientUpdated, err := h.s.updatePatient(id, reqPatient)
	if err != nil {
		if errors.Is(err, ErrorPatientNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, patientUpdated)
}
