package actions

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ArnaudCalmettes/microsocial/models"
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

// Full functional test scenario for friend requests & friendship
// Aka. The Short and Awkward Tale of Alice and Bob
func (as *ActionSuite) Test_Friendship() {
	// Set up users (admin, alice and bob)
	_, admin_token := as.createUserAndToken(true)
	alice, alice_token := as.createUserAndToken(false)
	bob, bob_token := as.createUserAndToken(false)

	nil_fr_url := fmt.Sprintf("/users/%s/friend_request", uuid.Nil)
	payload := &map[string]string{"message": "Let's be friends!"}

	// Unauthorized
	resp := as.JSON(nil_fr_url).Post(payload)
	as.Equal(401, resp.Code)

	// Alice requests friendship to her imaginary friend
	resp = as.createAuthRequest(nil_fr_url, alice_token).Post(payload)
	as.Equal(404, resp.Code) // ... But we can't find it

	alice_fr_url := fmt.Sprintf("/users/%s/friend_request", alice.ID)

	// Alice then grows some self-esteem and requests friendship to herself
	resp = as.createAuthRequest(alice_fr_url, alice_token).Post(payload)
	as.Equal(409, resp.Code) // ... But that's not how we do social.

	// Alice gives up: she obviously needs some help to make friends...

	// Bob requests friendship to Alice
	resp = as.createAuthRequest(alice_fr_url, bob_token).Post(payload)
	as.Equal(200, resp.Code)

	bob_alice_req := &models.FriendRequest{}
	err := json.Unmarshal(resp.Body.Bytes(), bob_alice_req)
	as.NoError(err)

	// Check that Bob can't spam Alice with friend requests
	resp = as.createAuthRequest(alice_fr_url, bob_token).Post(payload)
	as.Equal(409, resp.Code)

	//////////////////////////////////////////////////////////////////////
	// Visibility & privacy checks

	// Check that the incoming request appears in Alice's profile
	alice_profile := as.loadProfileAs(alice, alice_token)

	as.Empty(alice_profile.OutRequests, "Alice shouldn't have an outgoing request")
	as.NotEmptyf(alice_profile.InRequests, "Alice didn't receive the request")
	as.Equal(bob_alice_req.ID, alice_profile.InRequests[0].ID)
	as.Empty(alice_profile.InRequests[0].To, "'to' displayed in incoming request")
	as.NotEmptyf(alice_profile.InRequests[0].From, "Missing 'from' in incoming request")
	as.Equal(alice_profile.InRequests[0].From.ID, bob.ID)

	// Admin should also see it
	alice_profile = as.loadProfileAs(alice, admin_token)
	as.NotEmptyf(alice_profile.InRequests, "Admin can't see the incoming request")

	// Bob shouldn't see it
	alice_profile = as.loadProfileAs(alice, bob_token)
	as.Emptyf(alice_profile.InRequests, "Bob can see the request on Alice's profile")

	// Check that the outgoing request appears on Bob's profile
	bob_profile := as.loadProfileAs(bob, bob_token)
	as.Emptyf(bob_profile.InRequests, "Bob shouldn't have an incoming request")
	as.NotEmptyf(bob_profile.OutRequests, "Bob didn't receive the request")
	as.Equal(bob_alice_req.ID, bob_profile.OutRequests[0].ID)
	as.Empty(bob_profile.OutRequests[0].From, "'from' displayed in outgoing request")
	as.NotEmptyf(bob_profile.OutRequests[0].To, "Missing 'to' in outgoing request")
	as.Equal(bob_profile.OutRequests[0].To.ID, alice.ID)

	// Admin should also see it
	bob_profile = as.loadProfileAs(bob, admin_token)
	as.NotEmptyf(bob_profile.OutRequests, "Admin can't see the outgoing request")

	// Alice shouldn't see it
	bob_profile = as.loadProfileAs(bob, alice_token)
	as.Emptyf(bob_profile.OutRequests, "Alice can see the request on Bob's profile")

	///////////////////////////////////////////////////////////////////////////

	bob_fr_url := fmt.Sprintf("/users/%s/friend_request", bob.ID)

	// Check that Alice can't request friendship to Bob since there's already a
	// request from Bob to Alice. We're not Tinder: simply send a "clash" error
	// to avoid breaking the database's consistency.
	resp = as.createAuthRequest(bob_fr_url, alice_token).Post(payload)
	as.Equal(409, resp.Code)

	// Check that only Alice can accept or decline this friend request
	accept_url := fmt.Sprintf("/friend_requests/%s/accept", bob_alice_req.ID)
	decline_url := fmt.Sprintf("/friend_requests/%s/decline", bob_alice_req.ID)

	// Bob can't
	resp = as.createAuthRequest(accept_url, bob_token).Get()
	as.Equal(403, resp.Code)
	resp = as.createAuthRequest(decline_url, bob_token).Get()
	as.Equal(403, resp.Code)

	// Even Admin can't
	resp = as.createAuthRequest(accept_url, admin_token).Get()
	as.Equal(403, resp.Code)
	resp = as.createAuthRequest(decline_url, admin_token).Get()
	as.Equal(403, resp.Code)

	// But Alice can decline it.
	resp = as.createAuthRequest(decline_url, alice_token).Get()
	as.Equal(200, resp.Code)

	// Ooops! Actually Alice declined it by mistake,
	// so she tries to "accept it back"
	resp = as.createAuthRequest(accept_url, alice_token).Get()

	as.Equal(404, resp.Code) // ... But she can't, because the request is lost.

	// Check that the request disappeared from Alice and Bob's profiles
	alice_profile = as.loadProfileAs(alice, admin_token)
	as.Emptyf(alice_profile.InRequests, "Declined incoming request didn't disappear.")
	bob_profile = as.loadProfileAs(bob, admin_token)
	as.Emptyf(bob_profile.OutRequests, "Declined outgoing request didn't disappear.")

	// Check that Bob can make another friend request to Alice
	resp = as.createAuthRequest(alice_fr_url, bob_token).Post(payload)
	as.Equal(200, resp.Code)
	err = json.Unmarshal(resp.Body.Bytes(), bob_alice_req)
	as.NoError(err)

	accept_url = fmt.Sprintf("/friend_requests/%s/accept", bob_alice_req.ID)
	decline_url = fmt.Sprintf("/friend_requests/%s/decline", bob_alice_req.ID)

	// Alice accepts it
	resp = as.createAuthRequest(accept_url, alice_token).Get()
	as.Equal(200, resp.Code)

	// Now she can't decline it anymore
	resp = as.createAuthRequest(decline_url, alice_token).Get()
	as.Equal(404, resp.Code)

	///////////////////////////////////////////////////////////////////////////
	// Check friendship visibility

	alice_profile = as.loadProfileAs(alice, alice_token)
	as.NotEmptyf(alice_profile.Friends, "Alice should see she's friends with Bob")
	as.Equal(bob.ID, alice_profile.Friends[0].ID)

	bob_profile = as.loadProfileAs(bob, bob_token)
	as.NotEmptyf(bob_profile.Friends, "Bob should see he's friends with Alice")
	as.Equal(alice.ID, bob_profile.Friends[0].ID)

	bob_profile = as.loadProfileAs(bob, alice_token)
	as.Emptyf(bob_profile.Friends, "Alice shouldn't see Bob's friends")

	alice_profile = as.loadProfileAs(alice, admin_token)
	as.NotEmptyf(alice_profile.Friends, "Admin should see Alice's friends")

	// Alice hates herself, so she tries to unfriend herself
	alice_unfriend_url := fmt.Sprintf("/users/%s/unfriend", alice.ID)
	resp = as.createAuthRequest(alice_unfriend_url, alice_token).Get()
	as.Equal(400, resp.Code)

	// She can't, so she'll blame Bob and unfriend him
	bob_unfriend_url := fmt.Sprintf("/users/%s/unfriend", bob.ID)
	resp = as.createAuthRequest(bob_unfriend_url, alice_token).Get()
	as.Equal(200, resp.Code)

	///////////////////////////////////////////////////////////////////////////
	// Check that the friendship was successfully destroyed
	alice_profile = as.loadProfileAs(alice, alice_token)
	as.Emptyf(alice_profile.Friends, "Alice is still friends with Bob")

	bob_profile = as.loadProfileAs(bob, bob_token)
	as.Emptyf(bob_profile.Friends, "Bob is still friends with Alice")

	// If we reach here, then we can celebrate that Alice and Bob's very short
	// relationship made it through to its weird conclusion. Yay! \o/
}
