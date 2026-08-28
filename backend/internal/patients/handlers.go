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
