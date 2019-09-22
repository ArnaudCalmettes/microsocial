package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
)

// "Light" report model used by users during creation.
type LightReport struct {
	Info string `json:"info"` // Reason why the user is reported.
}

// Report model struct
type Report struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ByID      uuid.UUID `json:"-" db:"by_id"`
	By        *User     `json:"by,omitempty" db:"-" belongs_to:"user"`
	AboutID   uuid.UUID `json:"-" db:"about_id"`
	About     *User     `json:"about,omitempty" db:"-" belongs_to:"user"`
	Info      string    `json:"info" db:"info"`
}

// ReportFromLight Creates a full Report from its "light" version
func ReportFromLight(light *LightReport) *Report {
	return &Report{
		Info: light.Info,
	}
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
		&validators.UUIDIsPresent{Field: r.ByID, Name: "ByID"},
		&validators.UUIDIsPresent{Field: r.AboutID, Name: "AboutID"},
		&validators.StringIsPresent{Field: r.Info, Name: "Info"},
		&validators.FuncValidator{
			Field:   r.AboutID.String(),
			Name:    "AboutID",
			Message: "Can't report yourself (about_id: %s)",
			Fn:      func() bool { return r.ByID != r.AboutID },
		},
	), nil
}

// Create saves a newly created user into the database
func (r *Report) Create(tx *pop.Connection) (*validate.Errors, error) {
	return tx.ValidateAndCreate(r)
}
