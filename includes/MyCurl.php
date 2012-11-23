<?php
/**
 * MyCurl
 * Gets the contents of a specified URL. For our purposes, this is all that is needed to verify email sent.
 */
class MyCurl{
    // properties
    public $url;
    public $ch;

    /**
     * constructor
     * @param  string $url The URL whose contents we are going to get
     * @return void
     * Also, sets up basic curl setopt values for returning url content
     */
    function __construct($url){
        // init vars
        $this->url = $url;
        $this->ch  = curl_init();

        // set URL and other appropriate options
        curl_setopt($this->ch, CURLOPT_URL, $url);
        curl_setopt($this->ch, CURLOPT_HEADER, 0); 
        curl_setopt($this->ch, CURLOPT_RETURNTRANSFER, 1);
    }

    /**
     * exec
     * @return string Returns the content of the URL set in constructor
     */
    function exec(){
        return curl_exec($this->ch);
    }

    /**
     * destructor
     * @return void
     * Close the curl instance
     */
    function __destruct(){
        if ($this->ch) curl_close($this->ch);
    }
}
