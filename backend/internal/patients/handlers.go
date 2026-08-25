package patients

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var InMemoryPatients = []Patient{
	{
		Id:            "1",
		FirstName:     "Test",
		LastName:      "One",
		Birthday:      time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC),
		Phone:         "1234567890",
		Email:         "test@test.com",
		Status:        "Active",
		AdmissionDate: time.Now(),
	},
	{
		Id:            "2",
		FirstName:     "Test",
		LastName:      "Two",
		Birthday:      time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC),
		Phone:         "1234567890",
		Email:         "test@test.com",
		Status:        "Active",
		AdmissionDate: time.Now(),
	},
}

func getPatients(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, InMemoryPatients)
}
