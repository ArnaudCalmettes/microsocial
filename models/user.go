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
	ID          uuid.UUID      `json:"id" db:"id" fake:"skip"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at" fake:"skip"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at" fake:"skip"`
	Login       string         `json:"login" db:"login" fake:"{person.first}{person.last}"`
	Info        string         `json:"info" db:"info" fake:"{hipster.word}"`
	Friends     Users          `json:"friends,omitempty" db:"-"`
	OutRequests FriendRequests `json:"pending_requests,omitempty" db:"-" order_by:"created_at desc"`
	InRequests  FriendRequests `json:"incoming_requests,omitempty" db:"-" order_by:"created_at desc"`
	Reports     Reports        `json:"reports,omitempty" db:"-" order_by:"created_at desc"`
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

// SendRequest sends a friend request to given user
// This method is only used in tests. Use FriendRequest.Create() for finer
// error management.
func (u *User) SendRequest(tx *pop.Connection, to *User, msg string) (*FriendRequest, error) {
	req := &FriendRequest{
		FromID:  u.ID,
		ToID:    to.ID,
		Message: msg,
	}
	verrs, err := req.Create(tx)
	if err != nil {
		return nil, err
	}
	if verrs.HasAny() {
		return nil, verrs
	}
	return req, nil
}

// FetchFriends looks up a user's friends and fills its Friends list
func (u *User) FetchFriends(tx *pop.Connection) error {
	q := tx.Where("friendships.user_id = ?", u.ID)
	q = q.InnerJoin("friendships", "users.id = friendships.friend_id")
	q = q.Order("friendships.created_at desc")
	return q.All(&u.Friends)
}

// FetchOutRequests looks up a user's outgoing friend requests
func (u *User) FetchOutRequests(tx *pop.Connection) error {
	q := tx.Eager("To").Where("from_id = ? AND status = ?", u.ID, "PENDING")
	return q.All(&u.OutRequests)
}

// FetchInRequests looks up a user's incoming friend requests
func (u *User) FetchInRequests(tx *pop.Connection) error {
	q := tx.Eager("From").Where("to_id = ? AND status = ?", u.ID, "PENDING")
	return q.All(&u.InRequests)
}

// FetchRequests looks up a user's incoming and outgoing friend requests
func (u *User) FetchRequests(tx *pop.Connection) error {
	if err := u.FetchInRequests(tx); err != nil {
		return err
	}
	return u.FetchOutRequests(tx)
}

// FetchReports looks up any reports made about a user
func (u *User) FetchReports(tx *pop.Connection) error {
	return tx.Eager("By").Where("subject_id = ?", u.ID).All(&u.Reports)
}
