package patients

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getPatients(c *gin.Context) {
	patients, err := fetchPatients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusOK, patients)
}

func getPatient(c *gin.Context) {
	id := c.Param("id")
	patient, err := fetchPatient(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
	}
	if patient.FirstName == "" {
		c.JSON(http.StatusNotFound, "")
	}
	c.JSON(http.StatusOK, patient)
}

func addPatient(c *gin.Context) {
	var reqPatient PatientDB
	if err := c.BindJSON(&reqPatient); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
	}
	patient, err := createPatient(reqPatient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusCreated, patient)
}

func removePatient(c *gin.Context) {
	id := c.Param("id")
	patientId, err := deletePatient(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
	}
	if patientId == 0 {
		c.JSON(http.StatusNotFound, "")
	}
	c.JSON(http.StatusOK, patientId)
}

func changePatient(c *gin.Context) {
	var reqPatient Patient
	if err := c.BindJSON(&reqPatient); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
	}
	patientUpdated, err := updatePatient(reqPatient)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
	}
	c.JSON(http.StatusOK, patientUpdated)
}
