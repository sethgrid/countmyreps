<?php

/**
 * get totals
 * @param  array $data_array SQL results for a user or office
 * @return array             array with three keys: situps, pushups, pullups where value is the running total 
 * Due to crazy nesting of results and taking in two different arrangements of datasets, we have some nasty matching below
 */
function get_totals($data_array){
	#$return = array('situps'=>0, 'pushups'=>0, 'pullups'=>0);
	$return = array('burpees'=>0);
    foreach ($data_array as $k => $v){
        if (is_array($v)){
            foreach ($v as $k2 => $v2){
        		// now that we have dug into the array, we set our values based on two different
                // formats of array coming in. 
                #if ($k2 == 'situps' and is_numeric($v2)) $return['situps'] += $v2;
                if ($k2 == 'burpees' and is_numeric($v2)) $return['burpees'] += $v2;
                #else if ($k2 == 'pushups' and is_numeric($v2)) $return['pushups'] += $v2;
                #else if ($k2 == 'pullups' and is_numeric($v2)) $return['pullups'] += $v2;
                else{
                    $return['burpees'] += $v2['burpees'];
                    #$return['situps'] += $v2['situps'];
                    #$return['pushups'] += $v2['pushups'];
                    #$return['pullups'] += $v2['pullups'];
                }
            }
        }
    }
    return $return;
}
