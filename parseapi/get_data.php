<?php
include("../includes/DataStore.php");
include("../includes/func_get_totals.php");

$DataStore = new DataStore;
$content = '';

$user = $_POST['email'];
$user_id = $DataStore->user_exists($user);
if ($user_id){
    $header  = "<h3>Reps for $user</h3>";
    $info    = "";
    $content = "<table class='data'><th>Date/Time</th><th>Sit-ups</th><th>Push-ups</th><th>Pull-ups</th>";
    $all_records = $DataStore->get_all_records($user_id);

    foreach ($all_records as $date => $exercises_array){
        $content .= "<tr><td>$date</td>";
	// because of nasty nesting in data array, we have to make two passes to insure correct order of results
	foreach ($exercises_array as $index => $exercises){
	    foreach($exercises as $exercise => $rep){
		if ($exercise == "situps")  $situps  = $rep;
		if ($exercise == "pushups") $pushups = $rep;
		if ($exercise == "pullups") $pullups = $rep;
	    }
        }
        $content .= "<td>$situps</td><td>$pushups</td><td>$pullups</td>";
        $content .= "</tr>";
    }
    $cotent .= "</table>";

    // get total reps for the offices and the user
    $stats_california  = $DataStore->get_records_by_office("california");
    $stats_colorado    = $DataStore->get_records_by_office("colorado");

    $your_totals       = get_totals($all_records);
    $california_totals = get_totals($stats_california);
    $colorado_totals   = get_totals($stats_colorado);

    $info .= "Your Totals -- Situps: " . $your_totals['situps'] . 
             ", Pushups: " . $your_totals['pushups'] . 
             ", Pullups: " . $your_totals['pullups'] . 
             "<br />";
 
    $info .= "California Totals -- Situps: " . $california_totals['situps'] . 
             ", Pushups: " . $california_totals['pushups'] . 
             ", Pullups: " . $california_totals['pullups'] . 
             "<br />";
    
    $info .= "Colorado Totals -- Situps: " . $colorado_totals['situps'] . 
             ", Pushups: " . $colorado_totals['pushups'] . 
             ", Pullups: " . $colorado_totals['pullups'] . 
             "<br />";
}
else{
    $content = "No User Found";
}


// look up user
// if found, display reps

// else show error

?>
<html>
<head>
    <title>CountMyReps</title>
    <style  type="text/css">
        body{
            background-color: #E8E8E8;
        }
        div.center{
            margin: auto;
            background-color: white;
            width: 800px; 
            border: 1px solid #C8C8C8; 
            padding-top: 10px;
            padding-bottom: 20px;
       }
       div.inner{
            margin: auto;
            text-align: center;
            padding-top: 10px;
            color: #666362;
       }
       table.data{
	    margin: auto;
	    text-align: center;
	    border: 1px solid #C8C8C8;
	    color: #666362;
       }   
    </style>
</head>
<body>

<div class="center">
    <div class="inner">
    <?php
        echo $header.$info.$content;
    ?>
    </div>
</div>

</body>
</html>
