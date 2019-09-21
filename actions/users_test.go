package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ArnaudCalmettes/microsocial/models"
	"github.com/brianvoe/gofakeit"
	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"
)

// createUserAndToken creates a random user and generates the related auth
// token.
func (as *ActionSuite) createUserAndToken(is_admin bool) (*models.User, string) {
	user := as.createRandomUser()
	if is_admin {
		user.Admin = is_admin
		user.Update(as.DB)
	}
	token, err := newToken(user, time.Minute)
	as.NoError(err)
	return user, token
}

// loadProfileAs Queries user profile using given auth token
func (as *ActionSuite) loadProfileAs(user *models.User, token string) *models.User {
	resp := as.createAuthRequest(fmt.Sprintf("/users/%s", user.ID), token).Get()
	as.Equal(200, resp.Code)

	profile := &models.User{}
	err := json.Unmarshal(resp.Body.Bytes(), profile)
	as.NoError(err)

	return profile
}

func (as *ActionSuite) createRandomUser() *models.User {
	u := &models.User{}
	gofakeit.Struct(u)
	verrs, err := u.Create(as.DB)
	as.NoError(err)
	as.Falsef(verrs.HasAny(), verrs.String())
	return u
}

func (as *ActionSuite) createRandomUsers(n int) models.Users {
	users := make(models.Users, 0, n)
	for i := 0; i < n; i++ {
		u := as.createRandomUser()
		users = append(users, *u)
	}
	return users
}

func (as *ActionSuite) checkUsers(expected, actual models.Users) {
	for _, e := range expected {
		match := models.User{}
		for _, a := range actual {
			if e.ID == a.ID {
				match = a
			}
		}
		if match.ID == uuid.Nil {
			msg := fmt.Sprintf("User %v not in list", e)
			as.Fail(msg)
		}
		as.Equal(e.Login, match.Login)
		as.Equal(e.Info, match.Info)
	}
}

func (as *ActionSuite) createAuthRequest(url string, token string) *httptest.JSON {
	req := as.JSON(url)
	req.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)
	return req
}

func (as *ActionSuite) Test_Users_List() {
	expected := as.createRandomUsers(4)

	resp := as.JSON("/users").Get()
	as.Equal(200, resp.Code)

	actual := make(models.Users, 0, 4)
	err := json.Unmarshal(resp.Body.Bytes(), &actual)
	as.NoError(err)
	as.checkUsers(expected, actual)
}

func (as *ActionSuite) Test_Users_List_Paginated() {
	var results [10]models.User

	expected := as.createRandomUsers(10)

	// Get first page
	resp := as.JSON("/users?per_page=6").Get()
	as.Equal(200, resp.Code)

	users := models.Users(results[:])
	err := json.Unmarshal(resp.Body.Bytes(), &users)
	as.NoError(err)
	as.Equal(6, len(users))

	// Get the rest
	resp = as.JSON("/users?per_page=6&page=2").Get()
	as.Equal(200, resp.Code)

	users = models.Users(results[6:])
	err = json.Unmarshal(resp.Body.Bytes(), &users)
	as.NoError(err)
	as.checkUsers(expected, models.Users(results[:]))
}

func (as *ActionSuite) Test_Users_Show() {
	user := as.createRandomUser()
	other := as.createRandomUser()
	user_url := fmt.Sprintf("/users/%s", user.ID)
	nonexistent_url := "/users/non-existent"

	// Unauthorized
	resp := as.JSON(user_url).Get()
	as.Equal(401, resp.Code)

	// "Anonymous" token (anybody without admin credentials)
	token, err := newToken(other, time.Minute)
	as.NoError(err)
	resp = as.createAuthRequest(nonexistent_url, token).Get()
	as.Equal(404, resp.Code)

	resp = as.createAuthRequest(user_url, token).Get()
	as.Equal(200, resp.Code)

	actual := &models.User{}
	err = json.Unmarshal(resp.Body.Bytes(), &actual)
	as.NoError(err)
	as.Equal(user.ID, actual.ID)
	as.Equal(user.Login, actual.Login)
	as.Equal(user.Info, actual.Info)
}

func (as *ActionSuite) Test_Users_Create() {
	// Insufficient data
	req := as.JSON("/users")
	resp := req.Post(map[string]string{})
	as.Equal(409, resp.Code)

	// Valid request
	resp = req.Post(map[string]string{"login": "toto"})
	as.Equal(201, resp.Code)

	// Already exists
	resp = req.Post(map[string]string{"login": "toto"})
	as.Equal(409, resp.Code)
}

func (as *ActionSuite) Test_Users_Update() {
	var token string
	var err error

	user := as.createRandomUser()
	other := as.createRandomUser()
	admin := as.createRandomUser()
	admin.Admin = true
	admin.Update(as.DB)

	url := fmt.Sprintf("/users/%s", user.ID)

	// Unauthorized
	resp := as.JSON(url).Put(map[string]string{})
	as.Equal(401, resp.Code)

	// Use wrong, unprivileged user credentials
	token, err = newToken(other, time.Minute)
	as.NoError(err)
	req := as.createAuthRequest(url, token)
	resp = req.Put(map[string]string{})
	as.Equal(403, resp.Code)

	// Use authorized user credentials
	token, err = newToken(user, time.Minute)
	as.NoError(err)
	req = as.createAuthRequest(url, token)

	// No modification
	resp = req.Put(map[string]string{})
	as.Equal(200, resp.Code)

	// Modifying info
	resp = req.Put(map[string]string{"info": "Some offensive stuff"})
	as.Equal(200, resp.Code)
	actual := &models.User{}
	err = as.DB.Find(actual, user.ID)
	as.NoError(err)
	as.Equal("Some offensive stuff", actual.Info)

	// Trying to steal an existing login
	resp = req.Put(map[string]string{"login": other.Login})
	as.Equal(409, resp.Code)

	// Try to escalate privileges
	resp = req.Put(map[string]bool{"admin": true})
	as.Equal(403, resp.Code)

	// Use admin credentials
	token, err = newToken(admin, time.Minute)
	as.NoError(err)
	req = as.createAuthRequest(url, token)

	// Passivate the user's offensive information
	resp = req.Put(map[string]string{"info": "<Judge Dredd has pacified this info>"})
	as.Equal(200, resp.Code)
	err = as.DB.Find(actual, user.ID)
	as.NoError(err)
	as.Equal("<Judge Dredd has pacified this info>", actual.Info)

	// Promote user to admin
	resp = req.Put(map[string]bool{"admin": true})
	as.Equal(200, resp.Code)

	// Evil admin tries to change the user's login to another, existing one
	resp = req.Put(map[string]string{"login": other.Login})
	as.Equal(409, resp.Code)

	// Finally, try to modify a user that doesn't exist
	url = fmt.Sprintf("/users/%s", "non-existent")
	req = as.createAuthRequest(url, token)
	resp = req.Put(map[string]string{"info": "<Judge Dredd has pacified this info>"})
	as.Equal(404, resp.Code)
}

func (as *ActionSuite) Test_Users_Delete() {
	var token string
	var err error

	user := as.createRandomUser()
	other := as.createRandomUser()
	admin := as.createRandomUser()
	admin.Admin = true
	admin.Update(as.DB)

	url := fmt.Sprintf("/users/%s", user.ID)

	// Unauthorized
	resp := as.JSON(url).Delete()
	as.Equal(401, resp.Code)

	// Use wrong, unprivileged user credentials
	token, err = newToken(other, time.Minute)
	as.NoError(err)
	req := as.createAuthRequest(url, token)
	resp = req.Delete()
	as.Equal(403, resp.Code)

	// Use authorized user credentials
	token, err = newToken(user, time.Minute)
	as.NoError(err)
	req = as.createAuthRequest(url, token)

	// User deletes himself
	resp = req.Delete()
	as.Equal(200, resp.Code)

	// Use admin credentials
	token, err = newToken(admin, time.Minute)
	as.NoError(err)
	req = as.createAuthRequest(url, token)

	// User doesn't exist
	resp = req.Delete()
	as.Equal(404, resp.Code)

	url = fmt.Sprintf("/users/%s", other.ID)
	req = as.createAuthRequest(url, token)

	// Admin deletes the other user
	resp = req.Delete()
	as.Equal(200, resp.Code)
}
