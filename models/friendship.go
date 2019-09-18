package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
)

const (
	friendshipExists = `
		SELECT user_id
		FROM friendships
		WHERE user_id = ? AND friend_id = ?`
	friendRequestExists = `
		SELECT id
		FROM friend_requests
		WHERE status = ? AND from_id = ? AND to_id = ?`
)

// FriendRequest model struct
type FriendRequest struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	FromID    uuid.UUID `json:"-" db:"from_id"`
	ToID      uuid.UUID `json:"-" db:"to_id"`
	From      *User     `json:"from,omitempty" belongs_to:"user" `
	To        *User     `json:"to,omitempty" belongs_to:"user"`
	Message   string    `json:"message" db:"message"`
	Status    string    `json:"status" db:"status"`
}

// String is not required by pop and may be deleted
func (f FriendRequest) String() string {
	jf, _ := json.Marshal(f)
	return string(jf)
}

// FriendRequests is not required by pop and may be deleted
type FriendRequests []FriendRequest

// String is not required by pop and may be deleted
func (f FriendRequests) String() string {
	jf, _ := json.Marshal(f)
	return string(jf)
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
func (f *FriendRequest) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	var err error
	return validate.Validate(
		&validators.UUIDIsPresent{Field: f.FromID, Name: "FromID"},
		&validators.UUIDIsPresent{Field: f.ToID, Name: "ToID"},
		&validators.FuncValidator{
			Field:   f.ToID.String(),
			Name:    "ToID",
			Message: "Can't request friendship to yourself (%s)",
			Fn:      func() bool { return f.FromID != f.ToID },
		},
		&validators.FuncValidator{
			Field:   f.ToID.String(),
			Name:    "ToID",
			Message: "You're already friends with %s",
			Fn: func() bool {
				var id uuid.UUID
				q := tx.RawQuery(friendshipExists, f.FromID, f.ToID)
				b, err := q.Exists(&id)
				if err != nil {
					return false
				}
				return !b
			},
		},
		&validators.FuncValidator{
			Field:   f.ToID.String(),
			Name:    "ToID",
			Message: "There's already a pending friend request for %s",
			Fn: func() bool {
				var id uuid.UUID
				q := tx.RawQuery(friendRequestExists, "PENDING", f.FromID, f.ToID)
				b, err := q.Exists(&id)
				if err != nil {
					return false
				}
				return !b
			},
		},
	), err
}

func (f *FriendRequest) Create(tx *pop.Connection) (*validate.Errors, error) {
	f.Status = "PENDING"
	return tx.ValidateAndCreate(f)
}

func (f *FriendRequest) Accept(tx *pop.Connection) error {
	f.Status = "ACCEPTED"
	if err := tx.Update(f); err != nil {
		return err
	}
	return tx.RawQuery(
		"INSERT INTO friendships (user_id, friend_id) VALUES (?, ?), (?, ?)",
		f.FromID, f.ToID, f.ToID, f.FromID,
	).Exec()
}

func (f *FriendRequest) Decline(tx *pop.Connection) error {
	f.Status = "DECLINED"
	return tx.Update(f)
}
