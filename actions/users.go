package actions

import (
	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"
	"github.com/pkg/errors"
)

// UsersList lists all existing users
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

// UsersShow shows all available information about a user
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

// UsersCreate creates a new user
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
		return errors.WithStack(err)
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
