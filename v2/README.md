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

When the service starts up, it will create `./cmy.db` (a sqlite3 db) and initially populate it with exercise types and default teams. You can use a gui such as https://sqlitebrowser.org/ to browse data.

For Google Sign In, you will have to create a project on console.cloud.google.com. See https://skarlso.github.io/2016/06/12/google-signin-with-go/.
Create an OAauth2 Client ID and Secret. These are used to verifiy logged in users.

For local development, start the service binary with the env var `COUNTMY_REPS_DEV_MODE=true`. This will allow you to access APIs by allowing you to generate the bearer token without authenticating with Google. Hit `http://localhost:5000/auth?code=newuser@twilio.com` to generate a new bearer token.

```
18:20:51 sethammons@sethammons:~/workspace/repos/countmyreps/v2 (git:new-version*:c1be7d3)
$ curl 'localhost:5000/v3/token?code=seth.ammons@twilio.com'
{"Token":"FpxB2BfitcgMeb+IpCjZyv3uLnoUhSBmn-v+I4Z83HT_mrwJ"}
18:20:54 sethammons@sethammons:~/workspace/repos/countmyreps/v2 (git:new-version*:c1be7d3)
$ export TOKEN="FpxB2BfitcgMeb+IpCjZyv3uLnoUhSBmn-v+I4Z83HT_mrwJ"
18:21:09 sethammons@sethammons:~/workspace/repos/countmyreps/v2 (git:new-version*:c1be7d3)
$ curl localhost:5000/v3/stats -H "Authorization: Bearer $TOKEN" -d '{"Exercises":[{"Name":"Push Ups", "Count":15},{"Name":"Pull Ups", "Count": 5}]}'
18:21:23 sethammons@sethammons:~/workspace/repos/countmyreps/v2 (git:new-version*:c1be7d3)
$ curl localhost:5000/v3/stats -H "Authorization: Bearer $TOKEN"
[{"Date":"1586218883","Stats":[{"ID":1,"Name":"Push Ups","ValueType":"Reps","Count":15},{"ID":4,"Name":"Pull Ups","ValueType":"Reps","Count":5}]}]
```
### Compiling for Linux from Mac?

Because of the dependency on SQLite3 and due to issues with CGO and cross compilation, one cannot simply cross compile for linux from mac. Instead, the entire working directory needs to be loaded on a linux system with Go installed and compiled there.

```
scp -r * $LINUX_HOST:~/countmyreps
ssh $LINUX_HOST
cd countmyreps/v2/cmd/countmyreps
go build .
```

## API

### Authentication

Using Google oAuth2. A user clicks “sign in with google” and the callback url will be `{localhost:5000 | countmyreps.com}/auth` with args `state`, `code`, `scope`, `authuser`, `hd`, and `prompt`. Of these, we need to validate from the sign in page that the `state` we set (currently not exposed) is the `state` we get back. 

Then we pass the `code` value to the GET `/v3/token?code={:code:}` endpoint to receive the bearer token. This will be used in all authenticated requests as header `Authorization: Bearer {:token:}`. The token is good from 60 minutes or until server restart.

If the service was started with the environment variable `COUNTMYREPS_DEV_MODE=true`, then the `code` value will always validate successfully and store the value as the user's email address. I.E.: `?code=newuser@twilio.com`. 

#### `GET /v3/token`

`Options: ?code={:code:}`

Response:

```
{ “Token”: “some token” }
```

### Authenticated Endpoints

All the following endpoints require the header `Authorization: Bearer {:token:}`

#### `GET /v3/exercises`
Get All Exercise Options and their Type (“reps”, “km”, “minutes”, etc)

Response:
```
{
 “Exercises”: [{
    "ID": 3,
    “Name”: “Push Ups”,
    “Type”: “reps”
  }]
}
```

####`POST /v3/stats`

Submit reps

Post Body:
```
{
  “Exercises”:[{
    "ID": 3,            // ID is used if provided. If not provided, Name will be used to assign reps in the database.
    “Name”: “Push Ups”, // Name is used if ID is not provided
    “Count”: 15
  }]
}
```

Response:
201

#### `GET /v3/stats`
#### `GET /v3/stats/user/{:user_email:}`
#### `GET /v3/stats/team/{:team_id:}`
Options: `?startdate={:unix_ts:}&enddate={:unix_ts}`

Get the stats for all users, a particular user, or a particular team. Default start date is 31 days ago. Default end date is tomorrow. If there is any issue parsing the dates, they go to defaults silently.

Response:
```
{
  “Stats”:[{
    “Date”: “1586218883”,
    “Exercises”: [{
      "ID": 3,           // ID of the exercise
      “Name”: “Push Ups”,
      "ValueType": "reps"
      “Count”: 25
    }]
  }]
}
```

Create a Team
POST /v3/teams
{
  “Team”: “Irvine”
}
Resp:
201

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
Resp: 201

Leave a Team
DELETE /v3/myteams/id/{:team_id:}
Resp: 201

See what teams you are on
GET /v3/myteams
{
  “Teams”: [{
    “Name”: “Irvine”,
    “ID”: 4
  }]

