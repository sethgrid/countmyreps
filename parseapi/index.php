<?php

require_once("../includes/DataStore.php");        // connect to db
require_once("../includes/Logger.php");           // grab logger and log the incoming data
require_once("../includes/ReceiveParseAPI.php"); // grab the ReceiveParseAPI object
    
$Log = new Logger();
$Log->prefix("[parseapi post_data] ");
$Log->write(print_r($_POST, true));



// grab data
$Data = new ReceiveParseAPI($_POST);
if (!$Data->is_valid()) {
    // SendGrid's parse api will continually retry if we don't set status to 200
    header('HTTP/1.0 200 Successful', true, 200);
    $Log->write("Data failed validation - $Data->error");
    exit();
}

// post to db
$model = new DataStore;

if (!$model->user_exists($Data->from)){
    $model->create_user($Data->from);
}

$model->add_reps($Data->from, $Data->reps_hash);
// set response to 200
header('HTTP/1.0 200 Successful', true, 200);
$Log->write("Transaction Complete");
