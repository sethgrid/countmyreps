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
    function __construct($pdo_connection = null){
        $this->db_user = getenv("HTTP_DB_USER");
        $this->db_pass = getenv("HTTP_DB_PASS");
        $this->db_name = getenv("HTTP_DB_NAME");
        $this->db_host = 'localhost';

        // handle dependency injection
        if ($pdo_connection) $this->db = $pdo_connection;
        else $this->db = new PDO("mysql:host=$this->db_host;dbname=$this->db_name", $this->db_user, $this->db_pass);
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
        $now = date("Y-m-d H:i:s"); // not using SQL NOW() to avoid three inserts taking place upto a second appart
	    $query->bindParam(":date", $now);

        foreach ($rep_hash as $exercise=>$reps){
            $result = $query->execute();
        }

        if ($result){
            return true;
        }

        return false;
    }

    /**
     * get_count_by_office
     * @param  string $office The office the user is in
     * @return int            The count of users in that office
     */
    function get_count_by_office($office){
        $query = $this->db->prepare("SELECT COUNT(*) FROM `user` WHERE `office`=:office");
        $query->bindParam(":office", $office);
	    $query->execute();
        $result = $query->fetchAll();
        // the result is a multidimensional array, the first element on the first result is our count
	    return $result[0][0];
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
            $date     = $record['created_at'];
            $date     = explode(" ", $date);
            $date     = $date[0];
	        $today    = date('Y-m-d');
            $date_key = $date;

            // we want to show all of today's exercises by full time, everything else by day
	        if ($date == $today) $date_key = $record['created_at'];
                
            // initialize the key to avoid warnings
            #if (!array_key_exists($date_key, $return)) $return[$date_key] = array('situps'=>0, 'pushups'=>0, 'pullups'=>0);
            if (!array_key_exists($date_key, $return)) $return[$date_key] = array('burpees'=>0);
           
            // increment
            $return[$date_key][$record['exercise']] += $record['count'];
        }

	return $return;
    }

    /**
     * get_all_records_by_office
     * @param  int   $user_id the id of the user in the reps table
     * @return array          array indexed by date, second index by exercise, value is reps
     */
    function get_all_records_by_office($office){
        $query = $this->db->prepare("SELECT * FROM `reps` as r LEFT JOIN `user` as u on r.user_id=u.id WHERE u.office=:office ORDER BY r.created_at");
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

            #if (!array_key_exists($date, $return)) $return[$date] = array('situps'=>0, 'pushups'=>0, 'pullups'=>0);
            if (!array_key_exists($date, $return)) $return[$date] = array('burpees'=>0);
            $return[$date][$record['exercise']] += $record['count'];
        }

        return $return;
    }
} 
