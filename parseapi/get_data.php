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

    // to get the reps/person in each office, we need the total count of people in each office
    // there are two totals -- total people participating and total raw people
    $person_count_california_real = 34;
    $person_count_boulder_real    = 48;
    $person_count_denver_real     = 26;
    $person_count_nh_real         =  2;

    $person_count_california_participating = $DataStore->get_count_by_office('california');
    $person_count_boulder_participating    = $DataStore->get_count_by_office('boulder');
    $person_count_denver_participating     = $DataStore->get_count_by_office('denver');
    $person_count_nh_participating         = $DataStore->get_count_by_office('nh');

    $header  = "<h3>Reps for $user</h3>";

    $all_records_user       = $DataStore->get_all_records_by_user($user_id);
    $all_records_california = $DataStore->get_all_records_by_office('california');
    $all_records_boulder    = $DataStore->get_all_records_by_office('boulder');
    $all_records_denver     = $DataStore->get_all_records_by_office('denver');
    $all_records_nh         = $DataStore->get_all_records_by_office('nh');

    $user_reps_table       = format_as_table($all_records_user);
    $california_reps_table = format_as_table($all_records_california);
    $boulder_reps_table    = format_as_table($all_records_boulder);
    $denver_reps_table     = format_as_table($all_records_denver);
    $nh_reps_table         = format_as_table($all_records_nh);

    // get total reps for the offices and the user
    $your_totals       = get_totals($all_records_user);
    $california_totals = get_totals($all_records_california);
    $boulder_totals    = get_totals($all_records_boulder);
    $denver_totals     = get_totals($all_records_denver);
    $nh_totals         = get_totals($all_records_nh);

    $grand_total      = 0;
    $grand_total      = array_sum($california_totals); 
    $grand_total     += array_sum($boulder_totals); 
    $grand_total     += array_sum($denver_totals);
    $grand_total     += array_sum($hn_totals);

    $header .= 'Company total: ' . $grand_total;

    $info_u  = show_stats("Your",          $your_totals, 1, 1);
    $info_ca = show_stats("California",    $california_totals, $person_count_california_real, $person_count_california_participating);
    $info_co = show_stats("Boulder",       $boulder_totals,    $person_count_boulder_real,    $person_count_boulder_participating);
    $info_dn = show_stats("Denver",        $denver_totals,     $person_count_denver_real,     $person_count_denver_participating);
    $info_nh = show_stats("New Hampshire", $nh_totals,         $person_count_nh_real,         $person_count_nh_participating);

    if ($user == "none"){
	$info_u = '';
	$user_reps_table = '';
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
		<td class="cell"><?php echo $info_u.$user_reps_table;?></td>
		<td class="cell"><?php echo $info_ca.$california_reps_table;?></td>
		<td class="cell"><?php echo $info_co.$boulder_reps_table;?></td>
		<td class="cell"><?php echo $info_dn.$denver_reps_table;?></td>
		<td class="cell"><?php echo $info_nh.$nh_reps_table;?></td>
	</tr>
	</table>
    </div>
</div>

</body>
</html>
