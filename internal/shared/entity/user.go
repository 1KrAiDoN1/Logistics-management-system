package entity

import "time"

type User struct {
	ID                 int       `json:"id"`
	FirstName          string    `json:"first_name" validate:"required,min=2,max=50" example:"John"`
	LastName           string    `json:"last_name" validate:"required,min=2,max=50" example:"Doe"`
	Email              string    `json:"email" binding:"required"`
	Password           string    `json:"password" binding:"required"`
	TimeOfRegistration time.Time `json:"time_of_registration"`
}
