package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
)

// User model struct
type User struct {
	ID        uuid.UUID `json:"id" db:"id" fake:"skip"`
	CreatedAt time.Time `json:"created_at" db:"created_at" fake:"skip"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" fake:"skip"`
	Login     string    `json:"login" db:"login" fake:"{person.first}{person.last}"`
	Info      string    `json:"info" db:"info" fake:"{hipster.word}"`
}

// String converts a User to a JSON string
func (u User) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// Users is a collection of User.
type Users []User

// String converts a Users slice to a JSON string
func (u Users) String() string {
	ju, _ := json.Marshal(u)
	return string(ju)
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (u *User) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	var err error
	return validate.Validate(
		&validators.StringIsPresent{Field: u.Login, Name: "Login"},
		&validators.FuncValidator{
			Field:   u.Login,
			Name:    "Login",
			Message: "Login %s is already taken",
			Fn: func() bool {
				var b bool
				q := tx.Where("Login = ?", u.Login)
				if u.ID != uuid.Nil {
					q = q.Where("id = ?", u.ID)
				}
				b, err = q.Exists(u)
				if err != nil {
					return false
				}
				return !b
			},
		},
	), err
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
func (u *User) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	var err error
	return validate.Validate(
		&validators.FuncValidator{
			Field:   u.Login,
			Name:    "Login",
			Message: "Login %s is already taken",
			Fn: func() bool {
				var b bool
				q := tx.Where("login = ? AND id != ?", u.Login, u.ID)
				b, err = q.Exists(u)
				if err != nil {
					return false
				}
				return !b
			},
		},
	), err
}

// Create saves a newly created user into the database
func (u *User) Create(tx *pop.Connection) (*validate.Errors, error) {
	return tx.ValidateAndCreate(u)
}

// Update updates user information in the database
func (u *User) Update(tx *pop.Connection) (*validate.Errors, error) {
	return tx.ValidateAndUpdate(u)
}
