<?php

/**
 * format_as_table
 * @param  array  $all_records an array from the data model
 * @return string              an html table
 */
function format_as_table($all_records){
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
