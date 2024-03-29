# Microsocial

This is a toy social network-like REST API written in Golang, using Buffalo.

This project is actually a recruitement test. Requirements were to build an API allowing:

* Friend requests & friendship
* Reporting users to moderation
* Bonus: role-based visibility of information.

Given a time-frame of 10 hours (1 week's worth of spare time).

# Running image from Docker Hub

Using the following `docker-compose.yml` file:

```yaml
version: "3.7"
services:
    web:
        image: neuware/microsocial:latest
        environment:
        - JWT_SECRET=microsocial_secret
        - DATABASE_URL=postgres://buffalo:buffalo@db:5432/microsocial?sslmode=disable
        depends_on:
        - db
        ports:
        - "3000:3000"
    db:
        image: postgres:11
        environment:
        - POSTGRES_USER=buffalo
        - POSTGRES_PASSWORD=buffalo
        - POSTGRES_DB=microsocial
        ports:
        - "5432:5432"
```

* **During first launch**, first initialize the database by running `docker-compose up db`,
and wait until the DB is ready before interrupting it.
* Then, run `docker-compose up` to get the app up and running.

# Building from source

`docker-compose up` alone should do the trick, because the database will have plenty of time
to initialize as you build the app container. :)

# Discussing design choices

Given the time-frame (and the fact that I'm no Go expert), I followed a KISS approach and
strived for the simplest possible code to reach a 100% functional demo: no clever tricks,
no fancy stuff, only the couple requested features and bonus with a stress on *usability*.

Of course, this leaves room to improvement, and here are the three main things I think
should be addressed next in a "real world" scenario:

* **Friend requests are lost once accepted or declined**, because it makes it easier to
  prevent accepted/declined friend requests to be "declined/accepted back". A more realistic
  implementation would model this using a `status` column (`PENDING/ACCEPTED/DECLINED`) in the
  `friend_requests` table, in order to keep a logged record of all events.
* **Friendships are modelled using two distinct rows**, because, once again, it makes the rest of
  the code much simpler to read and write. As one could read in
[this StackOverflow thread](https://stackoverflow.com/questions/10807900/how-to-store-bidirectional-relationships-in-a-rdbms-like-mysql),
  there's actually a tie between "two rows, one index" or "one row, two indexes".
  The main drawback of the chosen path is that it could lead to inconsistent state in the database
  if we don't give it enough care: here, all "writes" or "deletes" in the `friendships` table
  manipulate **both rows** not only in the same transaction, but also **in the same query**, putting
  all the trust on the robustness of PostgreSQL transactions. IMO, it's the most reasonable choice,
  but I'm open to discussion and would gladly change my mind if convinced otherwise.
* Most importantly, **displaying a user's profile is the costliest action**, because it can yield up
  to 5 database queries to recover friend requests & friends (when a user visits his own profile), and
  moderation reports (when an admin visits a user profile). Note that these aren't the most likely
  use-case (the most likely case, "user A visits user B", takes only 1 DB query).
  I couldn't use "eager" fetches here (parameterizing the query as I'm checking the
  user's credentials), because each of the subrequests is already "eager" in its own right, in order
  to return human-friendly data. Given time and opportunity, I would advise to monitor how this performs
  "in real life", and if this operation were a bottleneck, my first approach wouldn't be to modify the
  code, but to start with an `If-Modified-Since` caching system that would benefit the whole API anyway
  (but that's the point of REST, isn't it?). I didn't do it here because, of course, there is nothing
  "small" or "easy" about caching, and it would be premature optimization anyway.

# Swagger documentation

When the app is launched, point your browser to `http://localhost:3000/swagger/index.html`
to display the API documentation.

# Step-by-step demo

If you prefer being told a functional narrative over reading a frozen cold
API documentation (I know I do...), you may find the following demo helpful. :)

## Setting up the stage

The app listens on TCP port 3000.

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

Authentication is basically out of the scope for this API, but there is a
need for some role management in the rest of it, so I wrote a simple "fake"
auth system based on JWT tokens.

If you place a GET request on the `/fake_auth/{login}` endpoint, you get a
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

Now, we can use `curl -H $AS_ALICE ...` to interact with the API on behalf
of Alice, and so on.

**If you're using swagger**, you can do the same by clicking the green *Authorize*
button and entering `Bearer <token value>` in the Value field. This will allow
you to run examples directly in the documentation interface.

Note: the default duration of the tokens is 24 hours. You can change this by
using an `exp` GET parameter when generating it. For instance:

* `?exp=10s` will generate a token that will expire in 10 seconds,
* `?exp=15m` for 15 minutes,
* `?exp=3h` for 3 hours.

## Basic user CRUD operations

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
$ curl -X PUT -H $AS_BOB -d '{"info": "Not a sponge"}' "http://$URL/users/$BOB_ID" | python3 -m json.tool
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

We'll simply quickly cover the nominal case here. A deeper and more thorough functional
test was written in `actions/friendships_test.go`.

Let's suppose Bob wants to become friends with Alice. He can place a friend request
by sending a POST to `/users/{user_id}/friend_request`.

```bash
$ curl -H $AS_BOB -d '{"message": "Please be my friend."}' \
    http://$URL/users/$ALICE_ID/friend_request
```

The request is visible as a "pending" request for Bob when he consults his profile,
and as an "incoming" request for Alice on her own profile:

```bash
$ curl -H $AS_BOB http://$URL/users/$BOB_ID |python3 -m json.tool
{
    "id": "d9e24321-cd55-4349-85f8-047bec35175c",
    "created_at": "2019-09-19T18:44:13.407875Z",
    "updated_at": "2019-09-19T20:32:36.739483Z",
    "login": "Bob",
    "info": "Not a sponge",
    "admin": false,
    "pending_requests": [
        {
            "id": "0f121386-f0cf-4f55-bb33-66a072d18801",
            "created_at": "2019-09-19T21:47:33.15393Z",
            "updated_at": "2019-09-19T21:47:33.153936Z",
            "to": {
                "id": "2acc6f8a-42ec-4f4b-bfe8-149ed0a83372",
                "created_at": "2019-09-19T18:43:52.715929Z",
                "updated_at": "2019-09-19T18:43:52.715934Z",
                "login": "Alice",
                "info": "Live from Wonderland",
                "admin": false
            },
            "message": "Please be my friend.",
        }
    ]
}
$ curl -H $AS_ALICE http://$URL/users/$ALICE_ID  |python3 -m json.tool
{
    "id": "2acc6f8a-42ec-4f4b-bfe8-149ed0a83372",
    "created_at": "2019-09-19T18:43:52.715929Z",
    "updated_at": "2019-09-19T18:43:52.715934Z",
    "login": "Alice",
    "info": "Live from Wonderland",
    "admin": false,
    "incoming_requests": [
        {
            "id": "0f121386-f0cf-4f55-bb33-66a072d18801",
            "created_at": "2019-09-19T21:47:33.15393Z",
            "updated_at": "2019-09-19T21:47:33.153936Z",
            "from": {
                "id": "d9e24321-cd55-4349-85f8-047bec35175c",
                "created_at": "2019-09-19T18:44:13.407875Z",
                "updated_at": "2019-09-19T20:32:36.739483Z",
                "login": "Bob",
                "info": "Not a sponge",
                "admin": false
            },
            "message": "Please be my friend.",
        }
    ]
}
```

Please note that this information is **private**, i.e. it is kept
invisible to everybody (including Bob) except Alice (and admins) on Alice's profile:

```bash
$ curl -H $AS_BOB http://$URL/users/$ALICE_ID |python3 -m json.tool
{
    "id": "2acc6f8a-42ec-4f4b-bfe8-149ed0a83372",
    "created_at": "2019-09-19T18:43:52.715929Z",
    "updated_at": "2019-09-19T18:43:52.715934Z",
    "login": "Alice",
    "info": "Live from Wonderland",
    "admin": false
}
```

Also, not shown here:

* Alice can't make a friend-request to Bob while there's
already one the other way round: this avoids breaking the consistency in
the database. A Tinder-like matchmaking system would be overkill here.
* Bob can't request himself as a friend,
* Bob can't make friend requests to his friends,
* Bob can have only one pending friend request to Alice:
he can make a second one only if she declines the first one.

Alice (and only her, not even admins) can either:

* Accept (`GET /friend_requests/{request_id}/accept`)
* Decline (`GET /friend_requests/{request_id}/decline`)

Let's accept it, and see that Bob and Alice are now friends:

```bash
$ curl -H $AS_ALICE http://$URL/friend_requests/0f121386-f0cf-4f55-bb33-66a072d18801/accept
$ curl -H $AS_ALICE http://$URL/users/$ALICE_ID | python3 -m json.tool
{
    "id": "2acc6f8a-42ec-4f4b-bfe8-149ed0a83372",
    "created_at": "2019-09-19T18:43:52.715929Z",
    "updated_at": "2019-09-19T18:43:52.715934Z",
    "login": "Alice",
    "info": "Live from Wonderland",
    "admin": false,
    "friends": [
        {
            "id": "d9e24321-cd55-4349-85f8-047bec35175c",
            "created_at": "2019-09-19T18:44:13.407875Z",
            "updated_at": "2019-09-19T20:32:36.739483Z",
            "login": "Bob",
            "info": "Not a sponge",
            "admin": false
        }
    ]
}
```

Once a friend request is accepted (resp. declined) it can't be "declined (resp. accepted) back".

Friendships are also private:

* Bob can't see Alice's friends, even if he's part of them.
* Admins can see Alice and Bob's friends.

Finally, Alice can unfriend Bob with `GET /users/{user_id}/unfriend`.

```bash
$ curl -H $AS_ALICE http://$URL/users/$BOB_ID/unfriend
"OK"
```

## Reporting users

Alice can report Bob to moderators, using the `/users/{user_id}/report` action. Note that in this
case she needs to provide information.

```bash
$ curl -H $AS_ALICE -d '{"info": "This user is a jerk"}' http://localhost:3000/users/$BOB_ID/report
```

Reports can the be listed by a `GET` call on the `/reports` endpoints, but they're obviously only
acressible by "admins" :

```bash
$ curl -H $AS_BOB http://localhost:3000/reports/
{"error":"Forbidden","status":403}
$ curl -H $AS_ADMIN http://localhost:3000/reports/ |python3 -m json.tool
[
    {
        "id": "25420fc8-5b42-4807-bac3-0313987add88",
        "created_at": "2019-09-21T18:05:14.049986Z",
        "by": {
            "id": "2acc6f8a-42ec-4f4b-bfe8-149ed0a83372",
            "created_at": "2019-09-19T18:43:52.715929Z",
            "updated_at": "2019-09-19T18:43:52.715934Z",
            "login": "Alice",
            "info": "Live from Wonderland",
            "admin": false
        },
        "about": {
            "id": "d9e24321-cd55-4349-85f8-047bec35175c",
            "created_at": "2019-09-19T18:44:13.407875Z",
            "updated_at": "2019-09-19T20:32:36.739483Z",
            "login": "Bob",
            "info": "Not a sponge",
            "admin": false
        },
        "info": "This user is a jerk"
    }
]
```

Alternatively, admins can also see reports made about a user when they're visiting this user's
profile:

```bash
$ curl -H $AS_ADMIN http://localhost:3000/users/$BOB_ID |python3 -m json.tool
{
    "id": "d9e24321-cd55-4349-85f8-047bec35175c",
    "created_at": "2019-09-19T18:44:13.407875Z",
    "updated_at": "2019-09-19T20:32:36.739483Z",
    "login": "Bob",
    "info": "Not a sponge",
    "admin": false,
    "reports": [
        {
            "id": "25420fc8-5b42-4807-bac3-0313987add88",
            "created_at": "2019-09-21T18:05:14.049986Z",
            "by": {
                "id": "2acc6f8a-42ec-4f4b-bfe8-149ed0a83372",
                "created_at": "2019-09-19T18:43:52.715929Z",
                "updated_at": "2019-09-19T18:43:52.715934Z",
                "login": "Alice",
                "info": "Live from Wonderland",
                "admin": false
            },
            "info": "This user is a jerk"
        }
    ]
}
```

Although this information stays invisible to "normal" users:

```bash
$ curl -H $AS_BOB http://localhost:3000/users/$BOB_ID |python3 -m json.tool
{
    "id": "d9e24321-cd55-4349-85f8-047bec35175c",
    "created_at": "2019-09-19T18:44:13.407875Z",
    "updated_at": "2019-09-19T20:32:36.739483Z",
    "login": "Bob",
    "info": "Not a sponge",
    "admin": false
}
```

