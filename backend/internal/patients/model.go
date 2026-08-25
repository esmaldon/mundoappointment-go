package patients

import (
	"time"
)

type Patient struct {
	Id            string    `json:"id"`
	FirstName     string    `json:"firstname"`
	LastName      string    `json:"lastname"`
	Birthday      time.Time `json:"birthday"`
	Phone         string    `json:"phone"`
	Email         string    `json:"email"`
	Status        string    `json:"status"`
	AdmissionDate time.Time `json:"admissionDate"`
}
