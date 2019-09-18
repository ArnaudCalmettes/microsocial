package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
)

// Report model struct
type Report struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	SubjectID uuid.UUID `json:"subject_id" db:"subject_id"`
	Message   string    `json:"message" db:"message"`
}

// String is not required by pop and may be deleted
func (r Report) String() string {
	jr, _ := json.Marshal(r)
	return string(jr)
}

// Reports is not required by pop and may be deleted
type Reports []Report

// String is not required by pop and may be deleted
func (r Reports) String() string {
	jr, _ := json.Marshal(r)
	return string(jr)
}

// Validate a Report
func (r *Report) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Field: r.UserID, Name: "UserID"},
		&validators.UUIDIsPresent{Field: r.SubjectID, Name: "SubjectID"},
		&validators.StringIsPresent{Field: r.Message, Name: "Message"},
		&validators.FuncValidator{
			Field:   r.SubjectID.String(),
			Name:    "SubjectID",
			Message: "Can't report yourself",
			Fn:      func() bool { return r.UserID != r.SubjectID },
		},
	), nil
}

// Create saves a newly created user into the database
func (r *Report) Create(tx *pop.Connection) (*validate.Errors, error) {
	return tx.ValidateAndCreate(r)
}
