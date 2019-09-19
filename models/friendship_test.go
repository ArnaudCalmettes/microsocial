package models

func (ms *ModelSuite) Test_FriendRequest_Create() {
	count, err := ms.DB.Count("friend_requests")
	ms.NoError(err)
	ms.Equal(0, count)

	user := ms.createRandomUser()
	other := ms.createRandomUser()

	request := &FriendRequest{
		FromID:  user.ID,
		ToID:    other.ID,
		Message: "Let's be friends!",
	}

	verrs, err := request.Create(ms.DB)
	ms.NoError(err)
	ms.Falsef(verrs.HasAny(), verrs.String())
	ms.NotZero(request.ID)
	ms.Equal("PENDING", request.Status)

	count, err = ms.DB.Count("friend_requests")
	ms.NoError(err)
	ms.Equal(1, count)

	err = ms.DB.Eager("From", "To").Find(request, request.ID)
	ms.NoError(err)
	ms.Equal(request.From.ID, user.ID)
	ms.Equal(request.To.ID, other.ID)

	err = user.FetchOutRequests(ms.DB)
	ms.NoError(err)
	ms.Equal(1, len(user.OutRequests))
	ms.Equal(other.ID, user.OutRequests[0].To.ID)
	ms.Nil(user.OutRequests[0].From)

	err = other.FetchInRequests(ms.DB)
	ms.NoError(err)
	ms.Equal(1, len(other.InRequests))
	ms.Equal(user.ID, other.InRequests[0].From.ID)
	ms.Nil(other.InRequests[0].To)

}

func (ms *ModelSuite) Test_FriendRequest_Accept() {
	user := ms.createRandomUser()
	other := ms.createRandomUser()

	request, err := user.SendFriendRequest(ms.DB, other, "")
	ms.NoError(err)

	err = request.Accept(ms.DB)
	ms.NoError(err)
	ms.Equal("ACCEPTED", request.Status)

	count, err = ms.DB.Count("friendships")
	ms.NoError(err)
	ms.Equal(2, count)

	err = user.FetchFriends(ms.DB)
	ms.NoError(err)
	ms.Equal(1, len(user.Friends))
	ms.Equal(other.ID, user.Friends[0].ID)

	err = other.FetchFriends(ms.DB)
	ms.NoError(err)
	ms.Equal(1, len(other.Friends))
	ms.Equal(user.ID, other.Friends[0].ID)
}

func (ms *ModelSuite) Test_FriendRequest_Decline() {
	user := ms.createRandomUser()
	other := ms.createRandomUser()

	request, err := user.SendFriendRequest(ms.DB, other, "")
	ms.NoError(err)

	err = request.Decline(ms.DB)
	ms.NoError(err)
	ms.Equal("DECLINED", request.Status)

	count, err := ms.DB.Count("friendships")
	ms.NoError(err)
	ms.Equal(0, count)
}

func (ms *ModelSuite) Test_FriendRequest_Validate() {
	user := ms.createRandomUser()
	other := ms.createRandomUser()

	request := &FriendRequest{
		FromID: user.ID,
		ToID:   user.ID,
	}

	verrs, err := request.Create(ms.DB)
	ms.NoError(err)
	ms.Truef(verrs.HasAny(), "Created friend request to self.")

	r1 := &FriendRequest{
		FromID: user.ID,
		ToID:   other.ID,
	}

	_, err = r1.Create(ms.DB)
	ms.NoError(err)

	r2 := &FriendRequest{
		FromID: user.ID,
		ToID:   other.ID,
	}

	verrs, err = r2.Create(ms.DB)
	ms.NoError(err)
	ms.Truef(verrs.HasAny(), "Created double friend request.")

	err = r1.Decline(ms.DB)
	ms.NoError(err)

	verrs, err = r2.Create(ms.DB)
	ms.NoError(err)
	ms.Falsef(verrs.HasAny(), "Couldn't create request when previous was declined.")

	err = r2.Accept(ms.DB)
	ms.NoError(err)

	r3 := &FriendRequest{
		FromID: user.ID,
		ToID:   other.ID,
	}
	verrs, err = r3.Create(ms.DB)
	ms.NoError(err)
	ms.Truef(verrs.HasAny(), "Created friend request to a friend.")
}
