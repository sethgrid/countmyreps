<?php
require_once ('includes/Logger.php');

class LoggerTest extends PHPUnit_Framework_TestCase{
    protected $file = "logs/test.log";

    /**
     * run before each test
     */
    protected function setUp(){
        // create a log fle
        $fh = fopen($this->file, "w");
        fclose($fh);
    }

    /**
     * tearDown
     * run after each test
     */
    protected function tearDown(){
        // remove the file
        unset($this->file);
    }
    /**
     * test that the logger will append to a file
     */
    public function testLoggerPrefixAndWrite(){
        // initialize variables
        $prefix  = '[TestLog]';
        $entry_1 = 'Entry One';
        $entry_2 = 'Entry Two';

        // set up logger and add two entries
        $Log = new Logger($this->file);
        $Log->prefix("[TestLog]");
        $Log->write("Entry One");
        $Log->write("Entry Two");

        // get file and set expectations
        $file         = file_get_contents($this->file);
        $expected_1   = $prefix . $entry_1;
        $expected_2   = $prefix . $entry_2;

        // date patter matches [2012-11-31 20:18:59]
        $date_pattern = "\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]";

        // match using regex because each entry is prepended with datetime and a space
        $this->assertRegExp("~$date_pattern " . preg_quote($expected_1) . "\n~", $file, "log file is as expected");
        $this->assertRegExp("~$date_pattern " . preg_quote($expected_2) . "\n~", $file, "log file is as expected");
    }
}
