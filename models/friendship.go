package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/gofrs/uuid"
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
}

type Friendship struct {
	CreatedAt time.Time `db:"created_at"`
	UserID    uuid.UUID `db:"user_id"`
	FriendID  uuid.UUID `db:"friend_id"`
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
	verrs := validate.Validate(
		&validators.UUIDIsPresent{Field: f.FromID, Name: "FromID"},
		&validators.UUIDIsPresent{Field: f.ToID, Name: "ToID"},
		&validators.FuncValidator{
			Field:   f.ToID.String(),
			Name:    "Self Friend",
			Message: "Can't request friendship to yourself (%s)",
			Fn:      func() bool { return f.FromID != f.ToID },
		},
	)
	if verrs.HasAny() {
		return verrs, nil
	}
	fs := []Friendship{}
	err := tx.Where("(user_id = ? AND friend_id = ?)", f.FromID, f.ToID).All(&fs)
	if err != nil {
		return verrs, err
	}
	if len(fs) > 0 {
		verrs.Add(
			"from_id/to_id",
			fmt.Sprintf(
				"%s and %s are already friends since %s",
				fs[0].UserID, fs[0].FriendID, fs[0].CreatedAt,
			),
		)
		return verrs, nil
	}
	fr := []FriendRequest{}
	err = tx.Where("(from_id = ? AND to_id = ?) OR (from_id = ? AND to_id = ?)",
		f.FromID, f.ToID, f.ToID, f.FromID).All(&fr)
	if err != nil {
		return verrs, err
	}
	if len(fr) > 0 {
		verrs.Add(
			"id",
			fmt.Sprintf(
				"There's already a pending friend request from %s to %s (id=%s)",
				fr[0].FromID, fr[0].ToID, fr[0].ID,
			),
		)
		return verrs, nil
	}
	return verrs, err
}

func (f *FriendRequest) Create(tx *pop.Connection) (*validate.Errors, error) {
	return tx.ValidateAndCreate(f)
}

func (f *FriendRequest) Accept(tx *pop.Connection) error {
	if err := tx.Destroy(f); err != nil {
		return err
	}
	fs := &Friendship{
		UserID:   f.FromID,
		FriendID: f.ToID,
	}
	return fs.Create(tx)
}

func (f *FriendRequest) Decline(tx *pop.Connection) error {
	return tx.Destroy(f)
}

// Create friendship
func (f *Friendship) Create(tx *pop.Connection) error {
	return tx.RawQuery(
		"INSERT INTO friendships (user_id, friend_id) VALUES (?, ?), (?, ?)",
		f.UserID, f.FriendID, f.FriendID, f.UserID,
	).Exec()
}

// Destroy friendship
func (f *Friendship) Destroy(tx *pop.Connection) error {
	return tx.RawQuery(
		`DELETE FROM friendships
		WHERE
			(user_id = ? AND friend_id = ?)
			OR
			(user_id = ? AND friend_id = ?)`,
		f.UserID, f.FriendID, f.FriendID, f.UserID,
	).Exec()
}
