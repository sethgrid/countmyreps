<?php
// modify the include path so that the required function that we are testing can load its dependencies
set_include_path("/home/seth/projects/countmyreps:/home/seth/projects/countmyreps/includes:" . get_include_path());

require_once ('includes/func_send_email_error.php');

class SendEmailErrorTest extends PHPUnit_Framework_TestCase{
    /**
     * test that the user's stats display shows correct totals
     */
    public function testCleanRun(){
        // mock logger will be called exactly once to log the response from CURL to the sendgrid web api  
        $MockLogger = $this->getMockBuilder('Logger')->disableOriginalConstructor()->setMethods(array('write', '__destruct'))->getMock();
        $MockLogger->expects($this->exactly(1))->method('write')->will($this->returnValue(true));
        
        $MockCurl   = $this->getMockBuilder('MyCurl')->disableOriginalConstructor()->setMethods(array('exec', '__destruct'))->getMock();
        $MockCurl->expects($this->exactly(1))->method('exec')->will($this->returnValue('{"message":"success"}'));


        $result = send_email_error(
                    $user_email   = "user@example.com", 
                    $addressed_to = "burpees@countmyreps.com", 
                    $subject      = "error email", 
                    $time         = "12:12:12 12-12-2012",
                    $Log          = $MockLogger,
                    $Curl         = $MockCurl
               );
        $expected = "There was an error with your CountMyReps Submission.\n" . 
                    "Make sure that you addressed your email to burpees@countmyreps.com.\n" . 
                    "Make sure that your subject line was one number, like: 24\n" . 
                    "\n" . 
                    "\n" . 
                    "Details from received message:\n" . 
                    "Addessed to: burpees@countmyreps.com\n" . 
                    "Subject: error email\n" . 
                    "Time: 12:12:12 12-12-2012\n";

        $this->assertEquals($expected, $result, 'Expected email content to be sent');
    }

    /**
     * test that Logger is called a second time if curl fails
     */
    public function testCurlFails(){
        // mock logger called exactly twice: once to log the response from curl, the second time to log the url that we hit
        // the second log write is triggered when curl fails
        $MockLogger = $this->getMockBuilder('Logger')->disableOriginalConstructor()->setMethods(array('write', '__destruct'))->getMock();
        $MockLogger->expects($this->exactly(2))->method('write')->will($this->returnValue(true));
        
        $MockCurl   = $this->getMockBuilder('MyCurl')->disableOriginalConstructor()->setMethods(array('exec', '__destruct'))->getMock();
        $MockCurl->expects($this->exactly(1))->method('exec')->will($this->returnValue('anything other than success json response'));


        $res = send_email_error(
                    $user_email   = "user@example.com", 
                    $addressed_to = "situps-pushups-pullups@countmyreps.com", 
                    $subject      = "error email", 
                    $time         = "12:12:12 12-12-2012",
                    $Log          = $MockLogger,
                    $Curl         = $MockCurl
               );
        // there are no assertions because the 'expects' call acts as a validator for the test
    }
}
