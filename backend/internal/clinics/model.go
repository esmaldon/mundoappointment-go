package clinics

import "uuid"

type Clinic struct {
	Id       *uuid.UUID `json:"id,omitempty"`
	Name     string     `json:"name"`
	Status   string     `json:"status"`
	Timezone string     `json:"timezone"`
}

type CreateClinic struct {
	Name     string `json:"name" binding:"required,min=1,max=50"`
	Timezone string `json:"timezone" binding:"required,min=1,max=50"`
}

type UpdateClinic struct {
	Name     *string `json:"name,omitempty" binding:"omitempty,min=1,max=50"`
	Status   *string `json:"status,omitempty" binding:"omitempty,oneof=Active Inactive"`
	Timezone *string `json:"timezone,omitempty" binding:"omitempty,min=1,max=50"`
}

func (u *UpdateClinic) IsEmpty() bool {
	return u.Name == nil &&
		u.Status == nil &&
		u.Timezone == nil
}
