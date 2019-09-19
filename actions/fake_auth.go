package actions

import (
	"time"

	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/pop"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

func newToken(u *models.User, exp time.Duration) (string, error) {
	claims := jwt.MapClaims{}
	claims["id"] = u.ID.String()
	claims["admin"] = u.Admin
	claims["exp"] = time.Now().Add(exp).Unix()
	secret, err := envy.MustGet("JWT_SECRET")
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func getCredentials(c buffalo.Context) *models.User {
	claims := c.Value("claims").(jwt.MapClaims)
	return &models.User{
		ID:    uuid.FromStringOrNil(claims["id"].(string)),
		Admin: claims["admin"].(bool),
	}
}

func LoginAsUser(c buffalo.Context) error {
	tx, ok := c.Value("tx").(*pop.Connection)
	if !ok {
		return errors.WithStack(errors.New("No transaction found"))
	}

	// Parse duration
	exp_str := c.Param("exp")
	if exp_str == "" {
		exp_str = "24h"
	}
	exp, err := time.ParseDuration(exp_str)
	if err != nil {
		return c.Error(400, err)
	}

	u := &models.User{}
	if err := tx.Where("login = ?", c.Param("login")).First(u); err != nil {
		return c.Error(404, err)
	}

	token, err := newToken(u, exp)
	if err != nil {
		return errors.WithStack(err)
	}

	return c.Render(200, r.JSON(token))
}
