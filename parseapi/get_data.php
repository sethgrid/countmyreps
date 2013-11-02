<?php
include("../includes/DataStore.php");
include("../includes/func_get_totals.php");
include("../includes/func_format_as_table.php");
include("../includes/func_show_stats.php");

$DataStore = new DataStore;
$content = '';
$header = 'Reps';
$display = array('california'=>null, 
		 'boulder'=>null,
		 'denver'=>null,
		 'nh'=>null,
		 'euro'=>null,
		 'other'=>null,);

$user = $_GET['email'];
$user_id = $DataStore->user_exists($user);
if ($user_id){

    $office_info = array( 
        array(
            'office' => 'california',
            'display_name' => 'Anaheim',
            'person_count' => 39,
        ),
        array(
            'office' => 'boulder',
            'display_name' => 'Boulder',
            'person_count' => 63,
        ),
        array(
            'office' => 'denver',
            'display_name' => 'Denver',
            'person_count' => 70,
        ),
        array(
            'office' => 'nh',
            'display_name' => 'Providence',
            'person_count' => 4,
        ),
        array(
            'office' => 'euro',
            'display_name' => 'Team Euro',
            'person_count' => 14,
        ),
	array(
	    'office' => 'other',
            'display_name' => 'Other',
	    'person_count' => 5,
        ),
    );

    $grand_total = 0;

    foreach ($office_info as $office){
        // creates a dynamic variable name such as $participating_california. Esp important for $display_$ofice_name
        $office_name = $office['office'];

        $participating = $DataStore->get_count_by_office($office_name);
        $all_records   = $DataStore->get_all_records_by_office($office_name);
        $reps_table    = format_as_table($all_records);
        $totals        = get_totals($all_records);
        $grand_total  += array_sum($totals);

        $stats                 = show_stats($office['display_name'], $totals, $office['person_count'], $participating);
        $display[$office_name] = $stats . '<br>' . $reps_table;
    }

    $all_records_user = $DataStore->get_all_records_by_user($user_id);
    $reps_table_user  = format_as_table($all_records_user);
    $totals_user      = get_totals($all_records_user);
    $stats_user       = show_stats("Your", $totals_user, 1, 1);
    $display_user     = $stats_user . '<br>' . $reps_table_user;

    $header  = "<h3>Reps for $user</h3>";
    $header .= 'Company total: ' . $grand_total . '<br><br><br>';

    if ($user == "none"){
        $display_user = '';
	    $header = "<h3>Reps for SendGrid</h3>";
    }
}
else{
    $header = "No User Found";
    $display = array();
    $display_user = '';
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
	    margin-left: 10px;
            background-color: white;
            width: 100%; 
            border: 1px solid #C8C8C8; 
            padding-top: 10px;
            padding-bottom: 20px;
       }
       div.inner{
            margin: auto;
	    margin-left: 10px;
            text-align: left;
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
            text-align: left;
	    padding-bottom: 50px;
            padding-top: 10px;
            color: #666362;
            vertical-align: top;
	    border-bottom: 1px solid gray;
       }
       table.icky{
           margin: auto;
	   margin-left: 10px;
       }
    </style>
</head>
<body>

<div class="center">
    <div class="inner">
        <?php
            echo $header;
        ?>
	<a href="#user">My Results</a> | <a href="#anaheim">Anaheim</a> | 
	<a href="#boulder">Boulder | <a href="#denver">Denver</a> | 
	<a href="#rhodeisland">Rhode Island</a> | <a href="#euro">Euro</a> |
	<a href="#other">Other</a><br><br> 
	<table class="icky">
	<tr>
		<td class="cell"><a name="user"></a><?php echo $display_user;?></td>
	</tr>
	<tr>
		<td class="cell"><a name="anaheim"></a><?php echo $display['california'];?></td>
	<tr>
	</tr>
		<td class="cell"><a name="boulder"></a><?php echo $display['boulder'];?></td>
	<tr>
	</tr>
		<td class="cell"><a name="denver"></a><?php echo $display['denver'];?></td>
	<tr>
	</tr>
		<td class="cell"><a name="rhodeisland"></a><?php echo $display['nh'];?></td>
	<tr>
	</tr>
		<td class="cell"><a name="euro"></a><?php echo $display['euro'];?></td>
	<tr>
	</tr>
		<td class="cell"><a name="other"></a><?php echo $display['other'];?></td>
	</tr>
	</table>
    </div>
</div>

</body>
</html>
