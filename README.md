# Microsocial

This is a toy social network-like REST API written in Golang, using Buffalo.

It is currently a Work In Progress.

# Step-by-step demo

The modelization is completely finished. I'm using this step-by-step demo as a guide
to develop the API.

## Setting up the stage

The app is currently usable in "development mode" (the Docker release will come soon).
To set it up, update "database.yml" to point to a valid PostgreSQL database with the right
credentials.

Let's start by defining a bunch of users, using POST requests on the `/users` endpoint:

```bash
$ URL="localhost:3000"
$ curl -d '{"login": "JudgeDredd", "admin": true, "info": "Pacificator Maximus"}' http://$URL/users/
$ curl -d '{"login": "Alice", "info": "Live from Wonderland"}' http://$URL/users/
$ curl -d '{"login": "Bob"}' http://$URL/users/
```

Edge cases:

* If a user login is already taken, you get a 409 error.

So we've defined an admin and two regular users. Let's list them:

```bash
$ curl http://$URL/users/ | python3 -m json.tool
[
    {
        "id": "9b9f01f4-34e7-4cc3-8837-b1ade476f72e",
        "created_at": "2019-09-19T18:42:53.971323Z",
        "updated_at": "2019-09-19T18:42:53.97133Z",
        "login": "JudgeDredd",
        "info": "Pacificator Maximus",
        "admin": true
    },
    {
        "id": "2acc6f8a-42ec-4f4b-bfe8-149ed0a83372",
        "created_at": "2019-09-19T18:43:52.715929Z",
        "updated_at": "2019-09-19T18:43:52.715934Z",
        "login": "Alice",
        "info": "Live from Wonderland",
        "admin": false
    },
    {
        "id": "d9e24321-cd55-4349-85f8-047bec35175c",
        "created_at": "2019-09-19T18:44:13.407875Z",
        "updated_at": "2019-09-19T18:44:13.407881Z",
        "login": "Bob",
        "info": "",
        "admin": false
    }
]
```

Notes:

* All list-like calls (such as this one) support pagination: they accept
`page` and `per_page` GET parameters, and return pagination detail in an
`X-Pagination` header.

... Aaaaand that's pretty much all we can do without authentication.

Let's save the IDs in env variables for later:

```bash
$ ADMIN_ID=9b9f01f4-34e7-4cc3-8837-b1ade476f72e
$ ALICE_ID=2acc6f8a-42ec-4f4b-bfe8-149ed0a83372
$ BOB_ID=d9e24321-cd55-4349-85f8-047bec35175c
```

## Authentication

Authentication is basically out of the scope for this API, but there is a need
for some basic role-management in the rest of it, so I wrote a simple
"fake" auth system based on JWT tokens.

Basically, if you place a GET request on the `/fake_auth/{login}` endpoint, you get a
token that you can pass as a `Authentication: Bearer <TOKEN>` header.

Let's define a couple env vars for the rest of this demo:

```bash
$ ADMIN_TOKEN=`curl http://$URL/fake_auth/JudgeDredd | sed 's/\"//g'`
$ ALICE_TOKEN=`curl http://$URL/fake_auth/Alice | sed 's/\"//g'`
$ BOB_TOKEN=`curl http://$URL/fake_auth/Bob | sed 's/\"//g'`
$ AS_ADMIN="Authorization: Bearer $ADMIN_TOKEN"
$ AS_ALICE="Authorization: Bearer $ALICE_TOKEN"
$ AS_BOB="Authorization: Bearer $BOB_TOKEN"
```

Now, we can use `curl -H $AS_ALICE ...` to interact with the API on behalf of Alice, and so
on.

Note: the default duration of the tokens is 24 hours. You can change this by using an `exp` GET
parameter when generating it. For instance:

* `?exp=10s` will generate a token that will expire in 10 seconds,
* `?exp=15m` for 15 minutes,
* `?exp=3h` for 3 hours.

# Basic user CRUD operations

The `/users` endpoint supports the classic CRUD operations:

* `GET /users` lists existing users (authentication not needed),
* `POST /users` creates a new user (authentication not needed),
* `GET /users/{user_id}` shows detailed user information, we'll see that later,
* `PUT /users/{user_id}` modifies an existing user,
* `DELETE /users/{user_id}` deletes a user,

We won't show everything here. Let's demo this by trying to change Bob's profile information.

```bash
$ curl -X PUT -d '{"info": "Not a sponge"}' "http://$URL/users/$BOB_ID"
{"error":"token not found in request","status":401}
```

Authentication is required. Let's try using Alice's token:

```bash
$ curl -X PUT -H $AS_ALICE -d '{"info": "Not a sponge"}' "http://$URL/users/$BOB_ID"
{"error":"Forbidden","status":403}
```

Of course, Alice can't modify Bob's information. Let's retry as Bob:

```bash
curl -X PUT -H $AS_BOB -d '{"info": "Not a sponge"}' "http://$URL/users/$BOB_ID"
{
    "id": "d9e24321-cd55-4349-85f8-047bec35175c",
    "created_at": "2019-09-19T18:44:13.407875Z",
    "updated_at": "2019-09-19T20:32:36.739483+02:00",
    "login": "Bob",
    "info": "Not a sponge",
    "admin": false
}
```

Success. This call can be used to modify:

* login (provided that the new login isn't already taken)
* info
* admin rights (although promotion requires admin credentials)

For instance, Bob can't escalate his own privileges:

```bash
curl -X PUT -H $AS_BOB -d '{"admin": true}' "http://$URL/users/$BOB_ID"
{"error":"I see what you did there!","status":403}
```

## Friends and friend requests

**To Be Continued...**
