CountMyReps
===========
Welcome to V2 of CountMyReps. Gone are the terrible days of PHP. I, for one, welcome our new Golang overlords.

### Running
There are required files in the directory. When you:
```
$ source sample.conf
$ go run main.go
```
You will need the `/web` and `/go_templates` directory and their contents relative to the running binary.

### Database
See `/setup` for a .sql file for setting up the database. There is one unexpected value that should be inserted into the office table: name = "". This allows us to leverage the empty string in Go and avoid null checks.

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
- refactor to be easier to work with
- consider how to do "team" grouping
- improve the UI/UX
- implement the JSON endpoint
- make a RESTful interface
- implement an integration test

Operability:
- move logs from `/var/log/messages` to their own location will rotation
- put monitoring and alerting on mariadb and countmyreps procs
- monitor for errors in the logs
- verify nightly back up of the db