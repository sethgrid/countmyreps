## CountMyReps Runbook

This document is here to help me remember next year how things are set up in production

### Hosting
The site is available on Digital Ocean

### Serving
The site is behind `nginx` (`/etc/nginx/conf.d`).

### Application
The application runs from `~/countmyreps`. There is a config file there that needs to be sourced and has all the config values.
The database is `mariaDB`, a `mysql` fork.

### Starting, Stopping
Leveraging `systemd` (`/etc/systemd/system/countmyreps.service`), we can `sudo service {countmyreps|mariadb} {start|stop}`. It takes care of the environment variables found in the config file.

### Logging
Currently, all logs to to syslog are are available in `/var/log/messages`. TODO: have countmyreps.log.

### Monitoring
The system uses `monit` (`/etc/monit.d/*`). Use `monit status {countmyreps|mysqld}`. Note that `mysqld` is actually `mariaDB`.