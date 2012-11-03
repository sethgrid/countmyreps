<?php

class DataStore
{
    # DataStore acts as a simple model. Allows for creating users, checking if they exist, and adding reps to the db

    private $db;
    private $db_user;
    private $db_pass;
    private $db_name;
    private $db_host;

    /**
     * constructor
     * @return void
     * Sets up the db connection, making use of environment variables that are set in .htaccess
     */
    function __construct(){
        $this->db_user = getenv("HTTP_DB_USER");
        $this->db_pass = getenv("HTTP_DB_PASS");
        $this->db_name = getenv("HTTP_DB_NAME");
        $this->db_host = 'localhost';

        $this->db = new PDO("mysql:host=$this->db_host;dbname=$this->db_name", $this->db_user, $this->db_pass);
    }

    /**
     * user_exists
     * @param  string $email The email address that is linked to the user adding reps
     * @return bool          Returns true is user already exists, false otherwise
     */
    function user_exists($email){
        $query = $this->db->prepare("SELECT * FROM `user` WHERE `email` = :email");
        $query->bindParam(":email", $email);
        $query->execute();

        if ($query->rowCount()){
            return true;
        }

        return false;
    }

    /**
     * create_user
     * @param  string $email The email address that is linked to the user adding reps
     * @return bool          Returns true if the user is added, false otherwise
     */
    function create_user($email){
        $query = $this->db->prepare("INSERT INTO `user`(`email`) VALUES (:email)");
        $query->bindParam(":email", $email);
        $result = $query->execute();

        if ($result){
            return true;
        }

        return false;
    }
}
