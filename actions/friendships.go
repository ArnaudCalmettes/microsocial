package actions

import (
	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
)

// FriendRequestsCreate places a new friend request.
func FriendRequestsCreate(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	auth := getCredentials(c)
	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(404, errors.New("Not Found"))
	}

	req := &models.FriendRequest{}
	if err := c.Bind(req); err != nil {
		return c.Error(400, err)
	}

	req.FromID = auth.ID
	req.ToID = user.ID

	verrs, err := req.Create(tx)
	if verrs.HasAny() {
		return c.Error(409, verrs)
	}
	if err != nil {
		return errors.WithStack(err)
	}

	return c.Render(200, r.JSON(req))
}

// FriendshipDestroy Unfriends a friend
func FriendshipsDestroy(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	auth := getCredentials(c)
	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(404, errors.New("Not Found"))
	}
	if auth.ID == user.ID {
		return c.Error(400, errors.New("Can't unfriend yourself"))
	}

	fs := &models.Friendship{
		UserID:   auth.ID,
		FriendID: user.ID,
	}
	if err := fs.Destroy(tx); err != nil {
		return errors.WithStack(err)
	}

	return c.Render(200, r.JSON("OK"))
}

// FriendRequestsAccept accepts a friend request.
func FriendRequestsAccept(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}
	req := &models.FriendRequest{}
	if err := tx.Find(req, c.Param("request_id")); err != nil {
		return c.Error(404, err)
	}

	auth := getCredentials(c)
	if req.ToID != auth.ID {
		return c.Error(403, errors.New("This request isn't yours to accept."))
	}

	if err := req.Accept(tx); err != nil {
		return errors.WithStack(err)
	}

	return c.Render(200, r.JSON(req))
}

// FriendRequestsDecline accepts a friend request.
func FriendRequestsDecline(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}
	req := &models.FriendRequest{}
	if err := tx.Find(req, c.Param("request_id")); err != nil {
		return c.Error(404, err)
	}
	auth := getCredentials(c)
	if req.ToID != auth.ID {
		return c.Error(403, errors.New("This request isn't yours to decline."))
	}

	if err := req.Decline(tx); err != nil {
		return errors.WithStack(err)
	}

	return c.Render(200, r.JSON(req))
}
