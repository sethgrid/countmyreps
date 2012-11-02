<?php

// connect to db
include("../includes/db_connect.php");

// grab logger and log the incoming data
include("../includes/logger.php");
$Log = new Logger("/var/tmp/logs/parseapi.log");
$Log->prefix("[parseapi post_data] ");
$Log->write(print_r($_POST, true);

// grab data

// verify

// post to db

// set response to 200

