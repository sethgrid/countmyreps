<?php

class Logger{
    /**
     * Simple logger class
     * Takes a filename in the constructor or defaults to logger.log
     */
    private $filename;
    private $fh;
    private $prefix;
    private $msg;

    /**
     * constructor
     * @param string $filename The file to which we will append.
     * @return void
     */
    function __construct($filename = ''){
        if ($filename){
		$this->filename = $filename;
        }
	else{
		// issue setting this as default in constructor
		$this->filename = '/home3/sethammo/www/countmyreps/logs/parseapi.log'; //getenv('HTTP_LOG_FILE');
	}
        $this->fh = fopen($this->filename, "a");
    }

    /**
     * prefix
     * @param string $prefix The prefix that prepends every log entry
     * @return void
     */
    function prefix($prefix = ""){
        $this->prefix = $prefix;
    }

    /**
     * write
     * @param string $message The message to append to the file.
     * @return void
     * A newline is always appended at the end of the $message
     */
    function write($message = ""){
	$now = date("Y-m-d H:i:s");
        fwrite($this->fh, '[' . $now . '] ' . $this->prefix . $message . "\n");
    }

    /**
     * destructor
     * close the file handle
     */
    function __destruct()
    {
        fclose($this->fh);
    }
}
