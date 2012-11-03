<?php

// connect to db
include("../includes/db_connect.php");

// grab logger and log the incoming data
include("../includes/logger.php");
$Log = new Logger("/var/tmp/logs/parseapi.log");
$Log->prefix("[parseapi post_data] ");
$Log->write(print_r($_POST, true));

// grab data
$to         = $_POST['to']; // TODO: situps-pushups-pullups@countmyreps.com will look in subject for 25, 20, 10 and assign accordingly
$from       = get_email($_POST['from']);
$reps_array = get_reps_array($_POST['subject']);

$Log->write("From: $from\nReps: $reps_array[0], $reps_array[1], $reps_array[2]");

// post to db

// set response to 200
http_response_code(200);

/**
 * get_email
 * @param  string $from_string the From Header
 * @return string              the email address
 * Expected input:
 *  - FirstName LastName <email@adder.ess>
 *  - <email@adder.ess>
 *  - email@adder.ess
 */ 
function get_email($from_string){
    // store matches in $matches. Being greedy, first result [0][0] will be "<email@addr.ess>". result [1][0] will be "email@addr.ess"
    preg_match_all('/<(.*)>/', $from_string, $matches);

    // does it kinda look like an email address?
    if (strstr($matches[1][0], "@")){
        return $matches[1][0];
    }
    else{
        return $from_string;
    }
}

/**
 * get_reps_array
 * @param  string $reps comma delimited string list of reps
 * @return array        array of values
 * TODO: change whole process to OO. Then make sure that element count matches to-string count
 */
function get_reps_array($reps){
    return explode(",", $reps);
}
