<?php

function format_as_table($all_records){
    $content = "<table class='data'><th>Date/Time</th><th>Sit-ups</th><th>Push-ups</th><th>Pull-ups</th>";
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

    return $content;
}
