package patients

type Patient struct {
	Id            int    `json:"id"`
	FirstName     string `json:"firstname"`
	LastName      string `json:"lastname"`
	Birthday      string `json:"birthday"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Status        string `json:"status"`
	AdmissionDate string `json:"admissionDate"`
}
