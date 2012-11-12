<?php

/**
 * format_as_table
 * @param  array  $all_records an array from the data model
 * @return string              an html table
 */
function format_as_table($all_records){

    // initialize empty array
    $init_array = Array();
    for ($d = 1; $d <= date('d'); $d++){
        // add leading zero if needed
	if ($d < 10) $d = "0$d";
	// make the key values equal to dates
	$init_array[date('Y-m-') . $d] = array('situps' => 0, 'pushups' => 0, 'pullups' => 0);
    }

    // merge to arrays together, over writing with records passed in
    $all_records = array_merge($init_array, $all_records);

    $content = "<table class='data'><th>Date/Time</th><th>Sit-ups</th><th>Push-ups</th><th>Pull-ups</th>";
    foreach ($all_records as $date => $exercises_array){
        $content .= "<tr><td>$date</td>";
	foreach ($exercises_array as $index => $rep){
	    if ($index == "situps")  $situps  = $rep;
	    if ($index == "pushups") $pushups = $rep;
	    if ($index == "pullups") $pullups = $rep;
        }
        $content .= "<td>$situps</td><td>$pushups</td><td>$pullups</td>";
        $content .= "</tr>";
    }
    $content .= "</table>";

    return $content;
}
