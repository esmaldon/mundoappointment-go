package patients

type Patient struct {
	Id            *int   `json:"id,omitempty"`
	FirstName     string `json:"firstname"`
	LastName      string `json:"lastname"`
	Birthday      string `json:"birthday"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Status        string `json:"status"`
	AdmissionDate string `json:"admissiondate"`
}

type CreatePatientRequest struct {
	FirstName string `json:"firstname" binding:"required,min=1,max=100"`
	LastName  string `json:"lastname" binding:"required,min=1,max=100"`
	Birthday  string `json:"birthday" binding:"required,datetime=2006-01-02"`
	Phone     string `json:"phone" binding:"required,min=10,max=15"`
	Email     string `json:"email" binding:"required,email"`
}

type UpdatePatientRequest struct {
	FirstName *string `json:"firstname,omitempty" binding:"omitempty,min=1,max=100"`
	LastName  *string `json:"lastname,omitempty" binding:"omitempty,min=1,max=100"`
	Birthday  *string `json:"birthday,omitempty" binding:"omitempty,datetime=2006-01-02"`
	Phone     *string `json:"phone,omitempty" binding:"omitempty,min=10,max=15"`
	Email     *string `json:"email,omitempty" binding:"omitempty,email"`
	Status    *string `json:"status,omitempty" binding:"omitempty,oneof=Active Inactive"`
}

func (u UpdatePatientRequest) IsEmpty() bool {
	return u.FirstName == nil &&
		u.LastName == nil &&
		u.Birthday == nil &&
		u.Phone == nil &&
		u.Email == nil &&
		u.Status == nil
}
