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
- Implement facebooks' Grace for graceful deployment

Operability:
- ~~verify nightly back up of the db~~ [done]
- move logs from `/var/log/messages` to their own location will rotation
- ~~put monitoring and alerting on mariadb and countmyreps procs~~ [done]
- monitor for errors in the logs
- implement db recover from logs (just in case, and if I have time)

### Docker
#### Source variables
We make pretty extensive use of environment variables here, it'd be a good idea to source them initially.  Feel free to update to your liking.

```bash
$ source sample.conf
```

#### Setup MySQL
Next up, setup database.

```bash
$ docker-compose build

$ docker-compose run -e MYSQL_ROOT_PASSWORD=my_awesome_password --entrypoint /opt/entrypoint.sh --u root --no-deps -T mysql
Creating countmyreps_data_1
root
>> mysql - waiting to become available
161106 05:13:06 mysqld_safe Logging to '/var/log/mysqld.log'.
161106 05:13:06 mysqld_safe Starting mysqld daemon with databases from /var/lib/mysql
>> mysql - not started
open
Warning: Using a password on the command line interface can be insecure.
Warning: Using a password on the command line interface can be insecure.
161106 05:13:08 mysqld_safe mysqld from pid file /var/run/mysqld/mysqld.pid ended
>> mysql - waiting to shutdown

$ docker-compose up -d mysql
Starting countmyreps_data_1
Creating countmyreps_mysql_1
```

#### Import Schema
Now that mysql is up and running, we need to create the initial database and tables.
```bash
$ docker-compose exec -T mysql mysql -u root -pmy_awesome_password -e "CREATE DATABASE IF NOT EXISTS $MYSQL_DBNAME;"
Warning: Using a password on the command line interface can be insecure.

$ docker-compose exec -T mysql mysql -u root -pmy_awesome_password -e "GRANT ALL PRIVILEGES ON $MYSQL_DBNAME.* to '$MYSQL_USER'@'%' IDENTIFIED BY '$MYSQL_PASS';"
Warning: Using a password on the command line interface can be insecure.

$ docker-compose -f docker-compose.yml -f docker-compose.setup.yml up schema
Starting countmyreps_data_1
countmyreps_mysql_1 is up-to-date
Recreating countmyreps_schema_1
Attaching to countmyreps_schema_1
countmyreps_schema_1 exited with code 0
```

#### Run service
Everything should now be set to run

```bash
$ docker-compose up countmyreps
Starting countmyreps_data_1
countmyreps_mysql_1 is up-to-date
Creating countmyreps_countmyreps_1
Attaching to countmyreps_countmyreps_1
countmyreps_1  | 2016/11/07 22:52:03 starting on :9126
```