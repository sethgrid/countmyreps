<?php
// modify the include path so that the required class that we are testing can load its dependencies
set_include_path("/home/seth/projects/countmyreps:/home/seth/projects/countmyreps/includes:" . get_include_path());
include ('includes/ReceiveParseAPI.php');

class ReceiveParseAPITest extends PHPUnit_Framework_TestCase{
    
    protected $raw_post;
    
    /**
     * setUp
     * run before each test, resets the default raw post
     */
    protected function setUp(){
        $this->raw_post = array(
                            'to'      => 'situps-pushups-pullups@countmyreps.com',
                            'from'    => '"Joe Sixpack" <joe@example.com>',
                            'subject' => '15',
                            'text'    => null,
                            'html'    => null,
                         );
    }

    /**
     * test that the user's stats display shows correct totals
     */
    public function testCleanRun(){
        $Parse = new ReceiveParseAPI($this->raw_post);
        $this->assertEquals(array(15),  $Parse->get('reps_array'));
    }

    /**
     * testInvalidTo
     * Change the raw post to have invalid to address. This makes the request invalid.
     */
    public function testInvalidTo(){
        $this->raw_post['to'] = 'invalid@countmyreps.com';
        $this->_assertInvalid();
    }

    /**
     * only comma separated delimeters should work
     */
    public function testInvalidSubjectDelimeter(){
        $this->raw_post['subject'] = '30/40/50';
        $this->_assertInvalid();
    }

    /**
     * test against missing value in subject
     */
    public function testInvalidSubjectMissingValue(){
        $this->raw_post['subject'] = '45, 15';
        $this->_assertInvalid();
    }

    /**
     * test that the getter works
     */
    public function testGetter(){
        $Parse = new ReceiveParseAPI($this->raw_post);
        
        $this->assertEquals('joe@example.com', $Parse->get('from'));
        $this->assertEquals(array(15),  $Parse->get('reps_array'));
    }

    /**
     * prior to running this method, in calling method set a raw_post value to something invalid.
     * mocked send_error to prevent that method from calling a function that will send an email.
     */
    private function _assertInvalid(){
        $mockParse = $this->getMock( $class_name          = 'ReceiveParseAPI', 
                                     $methods_to_mock     = array('send_error'), 
                                     $pass_to_constructor = array($this->raw_post)
                                   );
        $mockParse->expects($this->once())->method('send_error');
        
        $this->assertFalse($mockParse->is_valid(), "Correctly Invalid");
    }
}
