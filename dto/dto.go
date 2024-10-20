package dto

import (
	"time"

	"github.com/go-playground/validator"
)

type PostUser struct {
	Username string `json:"username" validate:"min=8,max=20"`
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"min=8,max=20"`
}

type GetUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type SetSeatsPriceDTO struct {
	SeatIDs []int `json:"seat_ids"`
	Price   int   `db:"price" json:"price" validate:"required,min=0"`
}

type PostEventDTO struct {
	Name        string    `json:"name" validate:"required,min=3,max=100"`
	StartTime   time.Time `json:"start_time" validate:"required"`
	EndTime     time.Time `json:"end_time" validate:"required,gtfield=StartTime"` // greater than StartTime, register validateEventTimes
	Status      string    `json:"status" validate:"required,oneof=draft scheduled ongoing finished cancelled"`
	VenueID     int       `json:"venue_id" validate:"required"`
	ArtistID    int       `json:"artist_id" validate:"required"`
	Description string    `json:"description,omitempty" validate:"max=500"`
}

// Custom validator to ensure EndTime is after StartTime
func validateEventTimes(fl validator.FieldLevel) bool {
	endTime := fl.Field().Interface().(time.Time)
	startTime := fl.Parent().FieldByName("StartTime").Interface().(time.Time)
	return endTime.After(startTime)
}

// RegisterCustomValidations registers custom validations for the validator
func RegisterCustomValidations(v *validator.Validate) {
	v.RegisterValidation("gtfield", validateEventTimes)
}
