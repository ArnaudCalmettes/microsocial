package actions

import (
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/pkg/errors"
)

func newToken(id string, is_admin bool, exp time.Duration) (string, error) {
	claims := jwt.MapClaims{}
	claims["id"] = id
	claims["admin"] = is_admin
	claims["exp"] = time.Now().Add(exp).Unix()
	secret, err := envy.MustGet("JWT_SECRET")
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func getCredentials(c buffalo.Context) (id string, is_admin bool) {
	claims, _ := c.Value("claims").(jwt.MapClaims)
	id, _ = claims["id"].(string)
	is_admin, _ = claims["admin"].(bool)
	return
}

// CreateToken creates a "fake" auth token
func CreateToken(c buffalo.Context) error {
	var id, admin, exp string
	if id = c.Param("id"); id == "" {
		id = "0" // Anonymous user by default
	}
	if admin = c.Param("admin"); admin == "" {
		admin = "0" // Non-admin by default
	}
	if exp = c.Param("exp"); exp == "" {
		exp = "15m"
	}

	is_admin, err := strconv.ParseBool(admin)
	if err != nil {
		return c.Error(400, err)
	}
	duration, err := time.ParseDuration(exp)
	if err != nil {
		return c.Error(400, err)
	}

	token, err := newToken(id, is_admin, duration)
	if err != nil {
		return errors.WithStack(err)
	}

	return c.Render(200, r.JSON(map[string]string{"token": token}))
}
