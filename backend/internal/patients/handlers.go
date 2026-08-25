package patients

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getPatients(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, fetchPatients())
}
