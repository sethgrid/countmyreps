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
     * @return int           Returns the user id or 0 for non-existent user
     */
    function user_exists($email){
        $query = $this->db->prepare("SELECT * FROM `user` WHERE `email` = :email");
        $query->bindParam(":email", $email);
        $query->execute();

        if ($query->rowCount()){
            $record = $query->fetch();
            return $record['id'];
        }

        return 0;
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

    /**
     * add_reps
     * @param  string $email    The email address associated to these reps we are adding
     * @param  array  $rep_hash Array keys are the exercise, values are the rep count
     * @return bool             True on success, false otherwise
     * Example rep_hash:
     * {'situps' => 36, 'pushups' => 24, 'pullups' => 12}
     */
    function add_reps($email, $rep_hash){
        // grap the user_id
        $user_id = $this->user_exists($email);

        // put the exercises and reps into the db
        $query = $this->db->prepare("INSERT INTO `reps` (`user_id`,`exercise`,`count`,`created_at`) VALUES (:user_id, :exercise, :count, :date)");
        $query->bindParam(":user_id", $user_id);
        $query->bindParam(":exercise", $exercise);
        $query->bindParam(":count", $reps);
	$query->bindParam(":date", date("Y-m-d H:i:s"));

        foreach ($rep_hash as $exercise=>$reps){
            $result = $query->execute();
        }

        if ($result){
            return true;
        }

        return false;
    }

    /**
     * get_all_records_by_user
     * @param  int   $user_id the id of the user in the reps table
     * @return array          array indexed by date, second index by exercise, value is reps
     */
    function get_all_records_by_user($user_id){
        $query = $this->db->prepare("SELECT * FROM `reps` WHERE `user_id`=:user_id ORDER BY `created_at`");
        $query->bindParam(":user_id", $user_id);
        $query->execute();
        $records = $query->fetchAll();

        $return = Array();
        // goal format for data:
        // Array [date][types of exercise] => reps for that day
        foreach ($records as $record){
            // get the date from the string (ie, text prior to the space)
            $date = $record['created_at'];
            $date = explode(" ", $date);
            $date = $date[0];
	    $today = date('Y-m-d');

            // we want to show all of today's exercises by full time, everything else by day
	    if ($date == $today){
                $return[$record['created_at']][$record['exercise']] += $record['count'];
            }
            else{
                $return[$date][$record['exercise']] += $record['count'];
           }
        }

        return $return;
    }

    /**
     * get_all_records_by_office
     * @param  int   $user_id the id of the user in the reps table
     * @return array          array indexed by date, second index by exercise, value is reps
     */
    function get_all_records_by_office($office){
        $query = $this->db->prepare("SELECT * FROM `reps`as r LEFT JOIN `user` as u on r.user_id=u.id WHERE u.office=:office ORDER BY r.created_at");
        $query->bindParam(":office", $office);
        $query->execute();
        $records = $query->fetchAll();

        $return = Array();
        // goal format for data:
        // Array [date][types of exercise] => reps for that day
        foreach ($records as $record){
            // get the date from the string (ie, text prior to the space)
            $date = $record['created_at'];
            $date = explode(" ", $date);
            $date = $date[0];
            $return[$date][$record['exercise']] += $record['count'];
        }

        return $return;
    }
    
    /**
     * get_records_by_office
     * @param  string $office The office of which the user is based (for office competitions)
     * @return array          Array indexed by date, second index by exercise, value is reps
     */
    function get_records_by_office($office){
        $query = $this->db->prepare("SELECT * FROM `reps` as r LEFT JOIN `user` as u on u.id=r.user_id WHERE `office`=:office ORDER BY `created_at`");
        $query->bindParam(":office", $office);
        $query->execute();
        $records = $query->fetchAll();
        $return = Array();
        foreach ($records as $record){
            $return[$record['created_at']][$record['exercise']] += $record['count'];
        }

        return $return;
    }
}
