<?php

$db_user = getenv("HTTP_DB_USER");
$db_pass = getenv("HTTP_DB_PASS");
$db_name = getenv("HTTP_DB_NAME");
$db_host = 'localhost';

$mysqlLink = mysql_connect($db_host, $db_user, $db_pass) or die ('Unable to connect to database. Please try again later.');
mysql_select_db($db_name);

