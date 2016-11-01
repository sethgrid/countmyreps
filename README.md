CountMyReps
===========
Welcome to V2 of CountMyReps. Gone are the terrible days of PHP. I, for one, welcome our new Golang overlords.

### Running
There are required files in the directory. When you:
```
$ source sample.conf
$ go build
$ ./countmyreps
```
You will need the `/web` and `/go_templates` directory and their contents relative to the running binary.

### Endpoints
```
    /                       # index; has form to input email to take you to /view
	/view                   # shows stats for you and the offices
	/json                   # json payload of the view page; good for anyone who wants to make a js frontend
	/healthcheck            # shows if the database is available
	/parseapi/index.php     # receives data from SendGrid's Inbound ParseAPI (legacy endpoint)
```
Anything in `/web` will be available via the file server. So `/web/images` will be available at `/images`.

### Database
See `/setup` for a .sql file for setting up the database. There is one unexpected value that should be inserted into the office table: name = "". This allows us to leverage the empty string in Go and avoid null checks.

Alternatively, you can set up and seed with some test data by running the integration test with:

`$ go test ./integration/... -overwrite-database -no-tear-down -mysql-dbname countmyreps`

### Simulating Inbound Parse Webhook
```
$ curl localhost:9126/parseapi/index.php -d to="pullups-pushups-squats-situps@countmyreps.com" -d from="someone@sendgrid.com" -d subject="1,2,3,4"
```

### Deploying
This is mostly just a note for me. Use `./build_n_upload.sh`.
This script will test, build, upload files, stop countmyreps, replace the binary, and start countmyreps.
One can check `/var/log/messages` to see log files being emitted. Remember to bump the version each deploy.

### TODO:

Code:
- ~~top navigation based on offices~~ [done]
- ~~implement the JSON endpoint~~ [done]
- make a RESTful interface
- implement an integration test
- refactor to be easier to work with
- consider how to do "team" grouping
- improve the UI/UX
- if we can't connect to the db, should the app try to recover the db with exec commands? Not good separation, but maybe good automation?
- `remote_addr` grabs nginx port; need real IP

Operability:
- ~~verify nightly back up of the db~~ [done]
- move logs from `/var/log/messages` to their own location will rotation
- ~~put monitoring and alerting on mariadb and countmyreps procs~~ [done]
- monitor for errors in the logs
- implement db recover from logs (just in case, and if I have time)