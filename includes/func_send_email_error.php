<?php

include ('sendgrid-php/SendGrid_loader.php');

/**
 * send_email_error
 * @param  string  $user_email   The user who sent the email to CountMyReps
 * @param  string  $addressed_to The CountMyReps email address that $user_email sent to
 * @param  string  $subject      The subject of the email sent in to CountMyReps
 * @param  string  $time         The time the email made it to CountMyReps
 * @return void
 * This function makes use of the SendGrid PHP library and uses their Web API rather than SMTP
 */
function send_email_error($user_email, $addressed_to, $subject, $time){
        // format $message to text/html value
        $msg  = "There was an error with your CountMyReps Submission.\n";
        $msg .= "Make sure that you addressed your email to situps-pushups-pullups@countmyreps.com.\n";
        $msg .= "Make sure that your subject line was three numbers separated by commas, like: 24, 12, 6\n";
        $msg .= "\n\n";
        $msg .= "Details from received message:\n";
        $msg .= "Addessed to: $addressed_to\n";
        $msg .= "Subject: $subject\n";
        $msg .= "Time: $time\n";

        $text_message = $msg;

        // setup and send the email to the user
        $sg_username = urlencode(getenv('HTTP_SG_USERNAME'));
        $sg_password = urlencode(getenv('HTTP_SG_PASSWORD'));
	$text_message = urlencode($text_message);

        // make a curl request over SendGrid's web api
        $url =  "https://sendgrid.com/api/mail.send.json?" . 
                "api_user=$sg_username" .
                "&api_key=$sg_password" . 
                "&to=$user_email" . 
                "&subject=$subject" . 
                "&text=$text_message" .
                "&from=error@countmyreps.com";

        // create a new cURL resource
        $ch = curl_init();

        // set URL and other appropriate options
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_HEADER, 0);

        // grab URL and pass it to the browser
        $result = curl_exec($ch);

        // close cURL resource, and free up system resources
        curl_close($ch);

	return;
}
