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
function send_email_error($user_email, $addressed_to, $subject, $time, $Log=null, $MyCurlRequest=null){
    
    // dependency injections
    if (!$Log){
        $Log = new Logger();
    }

	$Log->prefix("[send email] ");
	
	// format $message to text/html value
    $msg  = "There was an error with your CountMyReps Submission.\n";
    $msg .= "Make sure that you addressed your email to burpees@countmyreps.com.\n";
    $msg .= "Make sure that your subject line was one number, like: 24\n";
    $msg .= "\n\n";
    $msg .= "Details from received message:\n";
    $msg .= "Addessed to: $addressed_to\n";
    $msg .= "Subject: $subject\n";
    $msg .= "Time: $time\n";

    $text_message = $msg;
$text_message = 'Movember is over folks. Try again next year.';
    // setup and send the email to the user
    $sg_username = urlencode(getenv('HTTP_SG_USERNAME'));
    $sg_password = urlencode(getenv('HTTP_SG_PASSWORD'));
	
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
            "&from=error@countmyreps.com";
    
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
