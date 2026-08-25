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
