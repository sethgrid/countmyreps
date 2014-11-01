<?php
include("../includes/DataStore.php");
include("../includes/func_get_totals.php");
include("../includes/func_format_as_table.php");
include("../includes/func_show_stats.php");

$DataStore = new DataStore;
$content = '';
$header = 'Reps';
$display = array('oc'=>null, 
		 'boulder'=>null,
		 'denver'=>null,
		 'providence'=>null,
		 'euro'=>null,
		 'new_york'=>null,
		 'san_francisco'=>null,);

$user = $_GET['email'];
$user_id = $DataStore->user_exists($user);
if ($user_id){
    $user_office = str_replace("_", " ", $DataStore->get_location($user));
    if (emtpy(trim($user_office)){
    	$user_office = "unknown office - set your office location by sending an email with your office location in the subject";
    }
    $office_info = array( 
        'oc' => array(
            'office' => 'oc',
            'display_name' => 'oc',
            'person_count' => 53,
        ),
        'boulder' => array(
            'office' => 'boulder',
            'display_name' => 'Boulder',
            'person_count' => 74,
        ),
        'denver' => array(
            'office' => 'denver',
            'display_name' => 'Denver',
            'person_count' => 114,
        ),
        'providence' => array(
            'office' => 'providence',
            'display_name' => 'Providence',
            'person_count' => 8,
        ),
        'new_york' => array(
            'office' => 'new_york',
            'display_name' => 'New York',
            'person_count' => 3,
        ),
        'euro' => array(
            'office' => 'euro',
            'display_name' => 'Euro',
            'person_count' => 14,
        ),
	'san_francisco' => array(
	    'office' => 'san_francisco',
            'display_name' => 'San Francisco',
	    'person_count' => 4,
        ),
    );

    $grand_total = 0;

    foreach ($office_info as $office){
        $office_name = $office['office'];

        $participating = $DataStore->get_count_by_office($office_name);
        $all_records   = $DataStore->get_all_records_by_office($office_name);
        $reps_table    = format_as_table($all_records);
        $totals        = get_totals($all_records);
        $grand_total  += array_sum($totals);
        $stats                 = show_stats($office['display_name'], $totals, $office['person_count'], $participating);
        $display[$office_name] = $stats . '<br>' . $reps_table;

	# capture interesting data back into the info array
	# now the office_info array is useful if we made it json and available
	$office_info[$office_name]['totals'] = $totals;
	$office_info[$office_name]['total'] = array_sum($totals);
	$office_info[$office_name]['records'] = $all_records;
    }
    $all_records_user = $DataStore->get_all_records_by_user($user_id);
    $reps_table_user  = format_as_table($all_records_user);
    $totals_user      = get_totals($all_records_user);
    $stats_user       = show_stats("Your", $totals_user, 1, 1);
    $display_user     = $stats_user . '<br>' . $reps_table_user;

    # add the user to the office_info array for jsonifying
    $office_info['user'] = array(
    	'office' => '',
	'display_name' => $user,
	'person_count' => 1,
	'totals' => $totals_user,
	'total' => array_sum($totals_user),
	'records' => $all_records_user,
    );
    $header  = "<h3>Reps for $user ($user_office)</h3>";
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

if (@$_GET['json']){
   header('Content-Type: application/json');
   echo json_encode($office_info);
}
else{
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
	<a href="#user">My Results</a> | 
	<a href="#oc">OC (<?php echo $office_info['oc']['total'];?>)</a> | 
	<a href="#boulder">Boulder (<?php echo $office_info['boulder']['total'];?>)</a> | 
	<a href="#denver">Denver (<?php echo $office_info['denver']['total'];?>)</a> | 
	<a href="#providence">Providence (<?php echo $office_info['providence']['total'];?>)</a> |
	 <a href="#euro">Euro (<?php echo $office_info['euro']['total'];?>)</a> |
	<a href="#san_francisco">San Francisco (<?php echo $office_info['san_francisco']['total'];?>)</a> | 
	<a href="#new_york">New York (<?php echo $office_info['new_york']['total'];?>)</a> | 
	<a href=<?php echo "?email=".urlencode($user)."&json=1";?>>JSON</a><br><br> 
	<table class="icky">
	<tr>
		<td class="cell"><a name="user"></a><?php echo $display_user;?></td>
	</tr>
	<tr>
		<td class="cell"><a name="oc"></a><?php echo $display['oc'];?></td>
	<tr>
	</tr>
		<td class="cell"><a name="boulder"></a><?php echo $display['boulder'];?></td>
	<tr>
	</tr>
		<td class="cell"><a name="denver"></a><?php echo $display['denver'];?></td>
	<tr>
	</tr>
		<td class="cell"><a name="providence"></a><?php echo $display['providence'];?></td>
	<tr>
	</tr>
		<td class="cell"><a name="euro"></a><?php echo $display['euro'];?></td>
	<tr>
	</tr>
		<td class="cell"><a name="san_francisco"></a><?php echo $display['san_francisco'];?></td>
	</tr>
	</tr>
		<td class="cell"><a name="new_york"></a><?php echo $display['new_york'];?></td>
	</tr>
	</table>
    </div>
</div>

</body>
</html>
<?php } #endif ?>
