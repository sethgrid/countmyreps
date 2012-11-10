<?php
include ('func_send_email_error.php');
class ReceiveParseAPI
{
    /**
     * ReceiveParseAPI
     * grabs $_POST data from the SendGrid ParseAPI and maps some fields. Additionally, handles
     * rep counts based on subject and will validate the post data.
     */

    public $raw_post;
    public $to;
    public $from;
    public $subject;
    public $body;

    /**
     * constructor
     * @param  array $raw_post Takes in $_POST
     * @return void
     * TODO: have the to address specify the exercises that will be added (keys) and the subject have the values
     */
    function __construct($raw_post){
        $this->raw_post   = $raw_post;
        $this->to         = $raw_post['to'];
        $this->from       = $this->get_email($raw_post['from']);
        $this->subject    = $raw_post['subject'];
        $this->text       = $raw_post['text'];
        $this->html       = $raw_post['html'];
        $this->reps_array = $this->get_reps_array($raw_post['subject']);
        $this->reps_hash  = Array(
                              'situps'  => $this->reps_array[0], 
                              'pushups' => $this->reps_array[1], 
                              'pullups' => $this->reps_array[2]
                            );
    }

    /**
     * is_valid
     * @return bool Returns false if it is the wrong to address or if there are missing values for rep counts
     * Sends out an error message to the user who sent in an invalid email via email
     */
    function is_valid(){
        // make sure that we are receiving the right request
        if (!strstr($this->to, 'situps-pushups-pullups@countmyreps.com')){
	    $this->send_error();
            return false;
        }
        
        // make sure that we are getting values for all the reps
        foreach ($this->reps_hash as $exercise => $reps){
            if (!is_string($exercise) || !is_numeric($reps)){
                $this->send_error();
                return false;
            }
        }

        return true;
    }

    /**
     * lprint
     * @param string $msg a message to be printed to logs
     * @return void
     * Used for debugging the class if needed
     */
    function lprint($msg){
	require_once('Logger.php');
	$Logger = new Logger("/var/tmp/logs/debug.log");
	$Logger->write($msg);
    }

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

    /**
     * send_error
     * @return void
     */
    function send_error(){
	// from is the person who sent the email we recieved (ie, the sender who sent to us; the person to which we need to kick the error)
 	// to is the address they sent to (on our end)
	send_email_error($this->from, $this->to, $this->subject, date("Y-m-d H:i:s"));
	send_email_error(getenv('HTTP_MY_EMAIL'), $this->to, $this->subject . " (from $this->from)", date("Y-m-d H:i:s"));
    }
}
