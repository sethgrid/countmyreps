<?php

/*
$db_user = getenv("HTTP_DB_USER");
$db_pass = getenv("HTTP_DB_PASS");
$db_name = getenv("HTTP_DB_NAME");
$db_host = 'localhost';

$mysqlLink = mysql_connect($db_host, $db_user, $db_pass) or die ('Unable to connect to database. Please try again later.');
mysql_select_db($db_name);
*/

class DataStore
{
    private $db;
    private $db_user;
    private $db_pass;
    private $db_name;
    private $db_host;

    function __construct(){
        $this->db_user = getenv("HTTP_DB_USER");
        $this->db_pass = getenv("HTTP_DB_PASS");
        $this->db_name = getenv("HTTP_DB_NAME");
        $this->db_host = 'localhost';

        $this->db = new PDO("mysql:host=$this->db_host;dbname=$this->db_name", $this->db_user, $this->db_pass);
    }

    function user_exists($email){
        $query = $this->db_>prepare("SELECT * FROM `user` WHERE `email` = :email");
        $query->bindParam(":email", $email);
        $query->execute();

        if ($query->rowCount()){
            return true;
        }

        return false;
    }

    function create_user($email){
        $query = $this->db->prepare("INSERT INTO `user`(`email`) VALUES (:email)");
        $query->bindParam(":email", $email);
        $query->execute();

        if ($query->rowCount()){
            return true;
        }

        return false;
    }
}
