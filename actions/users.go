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
		return errors.WithStack(errors.New("no transaction found"))
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
		return errors.WithStack(errors.New("no transaction found"))
	}

	user := &models.User{}
	if err := tx.Find(user, c.Param("user_id")); err != nil {
		return c.Error(404, err)
	}
	return c.Render(200, r.JSON(user))

}

// UsersCreate creates a new user
func UsersCreate(c buffalo.Context) error {
	_, is_admin := getCredentials(c)

	if !is_admin {
		return c.Error(403, errors.New("Forbidden"))
	}

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	user := &models.User{}
	if err := c.Bind(user); err != nil {
		return errors.WithStack(err)
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
	id, is_admin := getCredentials(c)
	user_id := c.Param("user_id")

	if id != user_id && !is_admin {
		return c.Error(403, errors.New("Forbidden"))
	}

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	user := &models.User{}
	if err := tx.Find(user, user_id); err != nil {
		return c.Error(404, errors.New("Not Found"))
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
	id, is_admin := getCredentials(c)
	user_id := c.Param("user_id")

	if id != user_id && !is_admin {
		return c.Error(403, errors.New("Forbidden"))
	}

	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("no transaction found"))
	}

	user := &models.User{}

	if err := tx.Find(user, user_id); err != nil {
		return c.Error(404, err)
	}
	if err := tx.Destroy(user); err != nil {
		return errors.WithStack(err)
	}
	return c.Render(200, r.JSON(user))
}
