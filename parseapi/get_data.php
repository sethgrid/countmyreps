<?php
include("../includes/DataStore.php");
include("../includes/func_get_totals.php");
include("../includes/func_format_as_table.php");

$DataStore = new DataStore;
$content = '';

$user = $_GET['email'];
$user_id = $DataStore->user_exists($user);
if ($user_id){
    $header  = "<h3>Reps for $user</h3>";

    $all_records_user       = $DataStore->get_all_records_by_user($user_id);
    $all_records_california = $DataStore->get_all_records_by_office('california');
    $all_records_colorado   = $DataStore->get_all_records_by_office('colorado');

    $user_reps_table       = format_as_table($all_records_user);
    $california_reps_table = format_as_table($all_records_california);
    $colorado_reps_table   = format_as_table($all_records_colorado);

    // get total reps for the offices and the user
    $stats_california  = $DataStore->get_records_by_office("california");
    $stats_colorado    = $DataStore->get_records_by_office("colorado");

    $your_totals       = get_totals($all_records_user);
    $california_totals = get_totals($stats_california);
    $colorado_totals   = get_totals($stats_colorado);


    $info_u = "<p>Your Totals --  " . $your_totals['situps'] . 
             ", " . $your_totals['pushups'] . 
             ", " . $your_totals['pullups'] . 
             "<br /></p>";
 
    $info_ca = "<p>California Totals --  " . $california_totals['situps'] . 
             ", " . $california_totals['pushups'] . 
             ", " . $california_totals['pullups'] . 
             "<br /></p>";
    
    $info_co = "<p>Colorado Totals -- " . $colorado_totals['situps'] . 
             ", " . $colorado_totals['pushups'] . 
             ", " . $colorado_totals['pullups'] . 
             "<br /></p>";
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
            width: 1200px; 
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
       p{
            color: #666362;
            margin: auto;
       }
       table.data{
            display: inline;
	    margin: auto;
	    text-align: center;
	    border: 1px solid #C8C8C8;
	    color: #666362;
       }   
       td.cell{
            text-align: center;
            color: #666362;
       }
       table.icky{
           margin: auto;
       }
    </style>
</head>
<body>

<div class="center">
    <div class="inner">
        <?php
            echo $header;
        ?>
	<table class="icky">
	<tr>
		<td class="cell"><?php echo $info_u.$user_reps_table;?></td>
		<td class="cell"><?php echo $info_ca.$california_reps_table;?></td>
		<td class="cell"><?php echo $info_co.$colorado_reps_table;?></td>
	</tr>
	</table>
    </div>
</div>

</body>
</html>
