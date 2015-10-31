<?php
require_once ("Logger.php");
require_once ("MyCurl.php");

/**
 * send_email_error
 * @param  string  $user_email    The user who sent the email to CountMyReps
 * @param  string  $addressed_to  The CountMyReps email address that $user_email sent to
 * @param  string  $subject       The subject of the email sent in to CountMyReps
 * @param  string  $time          The time the email made it to CountMyReps
 * @param  Logger  $Log           Dependency Injection - Logger class
 * @param  MyCurl  $MyCurlRequest Dependency Injection - Curl class
 * @return void
 * This function makes use of the SendGrid PHP library and uses their Web API rather than SMTP
 */
function send_email_success($user_email, $addressed_to, $subject, $time, $Log=null, $MyCurlRequest=null){

    // dependency injections
    if (!$Log){
        $Log = new Logger();
    }

    $Log->prefix("[send email] ");
    $json = file_get_contents('http://countmyreps.com/parseapi/get_data.php?email='.urlencode($user_email).'&json=1');
    $data = json_decode($json, true);
    // format $message to text/html value
    $office = str_replace("_", " ", $data['user']['office']);
    $office_ = $data['user']['office'];

    $best_average_office_ = get_highest_average_office_name($data);
    $highest_average_office = str_replace("_", "", $best_average_office_);
        $average_reps_per_day = round(get_average_reps_per_day($best_average_office_, $data));

    $msg  = "Keep it up!\n";
    $msg .= "You've logged a total of ".$data['user']['total']." reps.\n";

    if ($office == ""){
        $msg .= "You do not have your office set. Send an email with your office name in the subject. Choose from oc, rwc, denver, boulder, providence, euro. This should be sent in ints own email.\n";
    } else {
        $percent_participating = round(100 * ((float)$data[$office_]['participating_count']/$data[$office_]['person_count']));
        $msg .= "The $office has a particpation rate of $percent_participating%. Currently, the $highest_average_office doing the best with $average_reps_per_day per person in the office per day.\n";
    }

    $msg .= "The totals are as follows: OC has ".$data['oc']['total'].", RWC has ".$data['rwc']['total'].", Boulder has ".$data['boulder']['total'].", Denver has ".$data['denver']['total'].", Providence has ".$data['providence']['total'].", New York has ".$data['new_york']['total'].", and the Euro team has ".$data['euro']['total'].".\n";

    $text_message = $msg;
    // setup and send the email to the user
    $sg_username = urlencode(getenv('HTTP_SG_USERNAME'));
    $sg_password = urlencode(getenv('HTTP_SG_PASSWORD'));


    $html_message = str_replace("\n","<br>",$text_message);
    $html_message = urlencode($html_message);
    $text_message = urlencode($text_message);
    $subject = urlencode($subject);
    if (!$subject) $subject = "missing";

    // make a curl request over SendGrid's web api
    $url =  "https://sendgrid.com/api/mail.send.json?" .
            "api_user=$sg_username" .
            "&api_key=$sg_password" .
            "&to=$user_email" .
            "&subject=" . trim($subject) .
            "&text=" . trim($text_message) .
            "&html=" . trim($html_message) .
            "&from=success@countmyreps.com";

    // dependency injection
    if (!$MyCurlRequest){
        $MyCurlRequest = new MyCurl($url);
    }

    // execute the curl request
    $result = $MyCurlRequest->exec();
    $Log->write($result);

    // if it was unsuccessful, log the result of the curl request for debugging
    if (!strstr($result, '{"message":"success"}')){
        $Log->write($url);
    }

    return $msg;
}

function get_highest_average_office_name($data){
    $office = "";
    $average = 0;
    foreach ($data as $office_name => $office_data){
        $tmp = $office_data["total"]/$office_data["person_count"];
        if ($tmp > $average && $office_name != "user"){
            $average = $tmp;
            $office = $office_name;
        }
    }
    return $office;
}

function get_average_reps_per_day($office_name_, $data){
    if ($office_name_ == ""){
        return 0;
    }
    $count = count($data[$office_name_]['records']);
    if ($count == 0) {
        $count = 1;
    }

    return $data[$office_name_]['total'] / $data[$office_name_]['person_count'] / $count;
}
