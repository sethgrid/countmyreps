<?php

class Logger{
    /**
     * Simple logger class
     * Takes a filename in the constructor or defaults to logger.log
     */
    private $filename;
    private $fh;
    private $msg;

    /**
     * constructor
     * @param string $filename The file to which we will append.
     * @return void
     */
    function __construct($filename = "logger.log"){
        $this->filename = $filename;
        $this->fh       = fopen($filename, "a");
    }

    /**
     * write
     * @param string $message The message to append to the file.
     * @return void
     * A newline is always appended at the end of the $message
     */
    function write($message = ""){
        fwrite($this->fh, $message . "\n");
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
