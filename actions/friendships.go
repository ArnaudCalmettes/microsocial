package actions

import (
	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
)

type LightFriendRequest struct {
	Message string `json:"message"`
}

// FriendRequestsCreate places a new friend request.
// @Summary Send a friend request to a user
// @Description Send a friend request to a user
// @security Bearer
// @Accept  json
// @Produce  json
// @Param user_id path string true "The user's ID"
// @Param message body actions.LightFriendRequest true "message associated to the friend request"
// @Success 200 {object} models.FriendRequest
// @Failure 400 {object} FormattedError
// @Failure 401 {object} FormattedError
// @Failure 404 {object} FormattedError
// @Failure 409 {object} FormattedError "You can't request yourself as a friend"
// @Router /users/{user_id}/friend_request [post]
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
// @Summary Unfriend another user
// @Description Unfriend another user
// @security Bearer
// @Produce  json
// @Param user_id path string true "The user's ID"
// @Success 200 {object} string
// @Failure 400 {object} FormattedError "Can't unfriend yourself"
// @Failure 401 {object} FormattedError
// @Failure 404 {object} FormattedError
// @Failure 500 {object} FormattedError
// @Router /users/{user_id}/unfriend [get]
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
// @Summary Accept a friend request
// @Description Accept a friend request
// @security Bearer
// @Produce  json
// @Param request_id path string true "The friend request ID"
// @Success 200 {object} models.FriendRequest
// @Failure 401 {object} FormattedError
// @Failure 403 {object} FormattedError "This request isn't yours to accept"
// @Failure 404 {object} FormattedError
// @Failure 500 {object} FormattedError
// @Router /friend_requests/{request_id}/accept [get]
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

// FriendRequestsDecline declines a friend request.
// @Summary Decline a friend request
// @Description Decline a friend request
// @security Bearer
// @Produce  json
// @Param request_id path string true "The friend request ID"
// @Success 200 {object} models.FriendRequest
// @Failure 401 {object} FormattedError
// @Failure 403 {object} FormattedError "This request isn't yours to decline"
// @Failure 404 {object} FormattedError
// @Failure 500 {object} FormattedError
// @Router /friend_requests/{request_id}/decline [get]
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
