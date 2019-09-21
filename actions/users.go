package actions

import (
	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
)

// UsersList lists all existing users
// @Summary List all users
// @Description List all existing users
// @Produce  json
// @Success 200 {object} models.Users
// @Header 200  {object} X-Pagination "pagination information"
// @Failure 500 {object} FormattedError
// @Router /users/ [get]
func UsersList(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("No transaction found"))
	}

	users := &models.Users{}

	// Paginate results. Params "page" and "per_page" control pagination.
	// Default values are "page=1" and "per_page=20".
	q := tx.PaginateFromParams(c.Params())

	if err := q.All(users); err != nil {
		return errors.WithStack(err)
	}

	// Add X-Pagination header
	c.Set("pagination", q.Paginator)
	return c.Render(200, r.JSON(users))
}

// UsersShow shows all available information about a user.
//
// This call may get costly with time (in the worst case scenario,
// it already performs 5 DB queries). However, this is a first iteration:
// let's see how it behaves in prod before optimizing anything.
//
// An obvious improvement may be to use a "If-Modified-Since" caching
// strategy.
// @Summary Show a user's profile
// @Description Show a detailed user profile.
// @Produce  json
// @security Bearer
// @Param user_id path string true "ID of the user"
// @Success 200 {object} models.User
// @Failure 401 {object} FormattedError
// @Failure 404 {object} FormattedError
// @Failure 500 {object} FormattedError
// @Router /users/{user_id} [get]
func UsersShow(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("No transaction found"))
	}

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(404, err)
	}

	auth := getCredentials(c)

	// Add extra (friends & friend requests) info
	if auth.ID == user.ID || auth.Admin {
		if err := user.FetchFriends(tx); err != nil {
			return c.Error(500, err)
		}
		if err := user.FetchRequests(tx); err != nil {
			return c.Error(500, err)
		}
	}

	// Add extra moderation info
	if auth.Admin {
		if err := user.FetchReports(tx); err != nil {
			return c.Error(500, err)
		}
	}

	return c.Render(200, r.JSON(user))

}

// Model accepted for user creation and modification
type LightUser struct {
	Login string `json:"login"` // User login (must be unique)
	Info  string `json:"info"`  // Optional user info
	Admin string `json:"admin"` // User has admin powers
}

// UsersCreate creates a new user
// @Summary Create a new user
// @Description Creates a new user
// @Accept  json
// @Produce  json
// @Param userinfo body actions.LightUser true "login (mandatory), info, admin"
// @Success 201 {object} models.User
// @Failure 400 {object} FormattedError
// @Failure 409 {object} FormattedError "The login is already taken"
// @Router /users/ [post]
func UsersCreate(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	user := &models.User{}
	if err := c.Bind(user); err != nil {
		return c.Error(400, err)
	}

	verrs, err := user.Create(tx)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		return c.Error(409, verrs)
	}

	return c.Render(201, r.JSON(user))
}

// UsersUpdate updates user information
// @Summary Update a user's information
// @Description Update a user's information
// @security Bearer
// @Accept  json
// @Produce  json
// @Param user_id path string true "The user ID"
// @Param userinfo body actions.LightUser true "New user information"
// @Success 200 {object} models.User
// @Failure 400 {object} FormattedError
// @Failure 401 {object} FormattedError
// @Failure 403 {object} FormattedError
// @Failure 409 {object} FormattedError
// @Router /users/{user_id} [put]
func UsersUpdate(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(404, errors.New("Not Found"))
	}

	auth := getCredentials(c)
	if auth.ID != user.ID && !auth.Admin {
		return c.Error(403, errors.New("Forbidden"))
	}

	if err := c.Bind(user); err != nil {
		return c.Error(400, err)
	}

	// Prevent users from escalating their own privileges.
	// Only admins can do that.
	if user.Admin && !auth.Admin {
		return c.Error(403, errors.New("I see what you did there!"))
	}

	verrs, err := user.Update(tx)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		return c.Error(409, verrs)
	}
	return c.Render(200, r.JSON(user))
}

// UsersDestroy deletes a user from the DB
// @Summary Deletes a user.
// @Description Deletes a user
// @security Bearer
// @Produce  json
// @Success 200 {object} models.User
// @Failure 401 {object} FormattedError
// @Failure 403 {object} FormattedError
// @Failure 404 {object} FormattedError
// @Router /users/{user_id} [delete]
func UsersDestroy(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(404, err)
	}

	auth := getCredentials(c)
	if auth.ID != user.ID && !auth.Admin {
		return c.Error(403, errors.New("Forbidden"))
	}

	if err := tx.Destroy(user); err != nil {
		return errors.WithStack(err)
	}
	return c.Render(200, r.JSON(user))
}
