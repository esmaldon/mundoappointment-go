package patients

import "uuid"

type Patient struct {
	Id            *uuid.UUID `json:"id,omitempty"`
	FirstName     string     `json:"firstname"`
	LastName      string     `json:"lastname"`
	Birthday      string     `json:"birthday"`
	Phone         string     `json:"phone"`
	Email         string     `json:"email"`
	Status        string     `json:"status"`
	AdmissionDate string     `json:"admissiondate"`
	ClinicId      uuid.UUID  `json:"clinic_id"`
	ParentId      *uuid.UUID `json:"parent_id"`
}

type Parent struct {
	Id        *uuid.UUID `json:"id,omitempty"`
	FirstName string     `json:"firstname"`
	LastName  string     `json:"lastname"`
	Birthday  string     `json:"birthday"`
	Status    string     `json:"status"`
}

type CreatePatientRequest struct {
	FirstName       string  `json:"firstname" binding:"required,min=1,max=100"`
	LastName        string  `json:"lastname" binding:"required,min=1,max=100"`
	Birthday        string  `json:"birthday" binding:"required,datetime=2006-01-02"`
	Phone           string  `json:"phone" binding:"required,min=10,max=15"`
	Email           string  `json:"email" binding:"required,email"`
	ParentFirstName *string `json:"parent_firstname,omitempty" binding:"omitempty,min=1,max=100"`
	ParentLastName  *string `json:"parent_lastname,omitempty" binding:"omitempty,min=1,max=100"`
	ParentBirthday  *string `json:"parent_birthday,omitempty" binding:"omitempty,datetime=2006-01-02"`
}

type UpdatePatientRequest struct {
	FirstName *string `json:"firstname,omitempty" binding:"omitempty,min=1,max=100"`
	LastName  *string `json:"lastname,omitempty" binding:"omitempty,min=1,max=100"`
	Birthday  *string `json:"birthday,omitempty" binding:"omitempty,datetime=2006-01-02"`
	Phone     *string `json:"phone,omitempty" binding:"omitempty,min=10,max=15"`
	Email     *string `json:"email,omitempty" binding:"omitempty,email"`
	Status    *string `json:"status,omitempty" binding:"omitempty,oneof=Active Inactive"`
}

type UpdateParentRequest struct {
	ParentFirstName *string `json:"parent_firstname,omitempty" binding:"omitempty,min=1,max=100"`
	ParentLastName  *string `json:"parent_lastname,omitempty" binding:"omitempty,min=1,max=100"`
	ParentBirthday  *string `json:"parent_birthday,omitempty" binding:"omitempty,datetime=2006-01-02"`
	ParentStatus    *string `json:"parent_status,omitempty" binding:"omitempty,oneof=Active Inactive"`
}

func (u *UpdatePatientRequest) IsEmpty() bool {
	return u.FirstName == nil &&
		u.LastName == nil &&
		u.Birthday == nil &&
		u.Phone == nil &&
		u.Email == nil &&
		u.Status == nil
}

func (u *UpdateParentRequest) IsEmpty() bool {
	return u.ParentFirstName == nil &&
		u.ParentLastName == nil &&
		u.ParentBirthday == nil &&
		u.ParentStatus == nil
}

func (c *CreatePatientRequest) IsParentEmpty() bool {
	return c.ParentFirstName == nil &&
		c.ParentLastName == nil &&
		c.ParentBirthday == nil
}

func (c *CreatePatientRequest) HasCompleteParent() bool {
	return c.ParentFirstName != nil && *c.ParentFirstName != "" &&
		c.ParentLastName != nil && *c.ParentLastName != "" &&
		c.ParentBirthday != nil && *c.ParentBirthday != ""
}
