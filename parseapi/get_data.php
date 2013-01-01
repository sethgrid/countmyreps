<?php
include("../includes/DataStore.php");
include("../includes/func_get_totals.php");
include("../includes/func_format_as_table.php");
include("../includes/func_show_stats.php");

$DataStore = new DataStore;
$content = '';
$hearder = 'Burpee Reps';

$user = $_GET['email'];
$user_id = $DataStore->user_exists($user);
if ($user_id){

    $office_info = array( 
        array(
            'office' => 'california',
            'display_name' => 'Anaheim';
            'person_count' => 34,
        ),
        array(
            'office' => 'boulder',
            'display_name' => 'Boulder';
            'person_count' => 48,
        ),
        array(
            'office' => 'denver',
            'display_name' => 'Denver';
            'person_count' => 26,
        ),
        array(
            'office' => 'new_hampshire',
            'display_name' => 'New Hampshire';
            'person_count' => 2,
        ),
        array(
            'office' => 'euro',
            'display_name' => 'Team Euro';
            'person_count' => 10,
        ),
    );

    $grand_total = 0;

    foreach ($office_info as $office){
        // creates a dynamic variable name such as $participating_california. Esp important for $display_$ofice_name
        $office_name = $office['office'];

        $participating_$office_name = $DataStore->get_count_by_office($office_name);
        $all_records_$office_name   = $DataStore->get_all_records_by_office($office_name);
        $reps_table_$office_name    = format_as_table($all_records_$office_name);
        $totals_$office_name        = get_totals($all_records_$office_name);
        $grand_total               += array_sum($totals_$office_name);

        $stats_$office_name   = show_stats($office['display_name'], $totals_$office_name, $participating_$office_name);
        $display_$office_name = $stats_$office_name . $reps_table_$office_name;
    }

    $all_records_user = $DataStore->get_all_records_by_user($user_id);
    $user_reps_table  = format_as_table($all_records_user);
    $totals_user      = get_totals($all_records_user);
    $stats_user       = show_stats("Your", $your_totals, 1, 1);
    $display_user     = $stats_user . $totals_user;

    $header  = "<h3>Reps for $user</h3>";
    $header .= 'Company total: ' . $grand_total;

    if ($user == "none"){
        $display_user = '';
	    $header = "<h3>Reps for SendGrid</h3>";
    }
}
else{
    $content = "No User Found";
}

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
            width: 1250px; 
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
            vertical-align: top;
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
		<td class="cell"><?php echo $display_user;?></td>
		<td class="cell"><?php echo $display_california;?></td>
		<td class="cell"><?php echo $display_boulder;?></td>
		<td class="cell"><?php echo $display_denver;?></td>
		<td class="cell"><?php echo $display_new_hampshire;?></td>
	</tr>
	</table>
    </div>
</div>

</body>
</html>
