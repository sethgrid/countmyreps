# CountMyReps

## Local Development

You will need to install sqlite3 and go-sqlite3. Using Go 1.14 and vendoring:
```
GOPROXY=direct go get github.com/mattn/go-sqlite3
go mod vendor
go install github.com/mattn/go-sqlite3
```

Now you can build and run the local binary:
```
cd ./cmd/countmyreps
go build .
./countmyreps
```

When the service starts up, it will create `./cmy.db` (a sqlite3 db) and initially populate it with exercise types and default teams.


For Google Sign In, you will have to create a project on console.cloud.google.com. See https://skarlso.github.io/2016/06/12/google-signin-with-go/.
Create an OAauth2 Client ID and Secret. These are used to verifiy logged in users.

For local development, start the service binary with `-local-dev=true`. This will allow you to access APIs normally blocked by requiring the X-GOOGLE-TOKEN header. 

### Compiling for Linux from Mac?

Because of the dependency on SQLite3 and due to issues with CGO and cross compilation, one cannot simply cross compile for linux from mac. Instead, the entire working directory needs to be loaded on a linux system with Go installed and compiled there.

```
scp -r * $LINUX_HOST:~/countmyreps
ssh $LINUX_HOST
cd countmyreps/v2/cmd/countmyreps
go build .
```

## API

#### Authentication
Using Google oAuth2. A user clicks “sign in with google” and the callback url will be `{localhost:5000 | countmyreps.com}/auth` with args `state`, `code`, `scope`, `authuser`, `hd`, and `prompt`. Of these, we need to validate from the sign in page that the `state` we set (currently not exposed) is the `state` we get back. 

Then we pass the `code` value to the GET `/v3/token?code={:code:}` endpoint to receive the bearer token. This will be used in all authenticated requests as header `Authorization: Bearer {:token:}`. The token is good from 60 minutes or until server restart

`GET /v3/token`

`Options: ?code={:code:}`

Response:

```
{ “Token”: “some token” }
```

All the following endpoints require the header `Authorization: Bearer {:token:}`


Get All Exercise Options and their Type (“reps”, “km”, “minutes”, etc)

`GET /v3/exercises`

Response:
```
{
 “Exercises”: [{
    “Name”: “Push Ups”,
    “Type”: “reps”
  }]
}
```

Get the stats for all users, a particular user, or a particular team

`GET /v3/stats`

`/v3/stats/user/{:user_email:}`

`/v3/stats/team/{:team_id:}`

Options: `?startdate={:date_1:}&enddate={:date_2}`

Response:
```
{
  “Stats”:[{
    “Date”: “2020-04-03”,
    “Exercises”: [{
      “Name”: “Push Ups”,
      “Count”: 25
    }]
  }]
}
```

Submit reps
POST /v3/stats
{
  “Exercises”:[{
    “Name”: “Push Ups”,
    “Count”: 15
  }]
}
Response:
204

Create a Team
POST /v3/teams
{
  “Team”: “Irvine”
}
Resp:
204

Delete a Team (maybe not expose this one? Limit its access to its creator?)
DELETE /v3/teams
{
  “TeamID”: 4
}

Get teams (all teams, not the signed in user’s teams)
GET /v3/teams
{
  “Teams”: [{
    “Name”: “Irvine”,
    “ID”: 4
  }]
}

Join a Team
POST /v3/myteams/id/{:team_id:}
Resp: 204

Leave a Team
DELETE /v3/myteams/id/{:team_id:}
Resp: 204

See what teams you are on
GET /v3/myteams
{
  “Teams”: [{
    “Name”: “Irvine”,
    “ID”: 4
  }]

