<?php

/**
 * show_stats
 * @param  string $who                      Whos totals are we displaying?
 * @param  array  $totals                   Expected Keys: situps, pushups, pullups; values are rep counts
 * @param  int    $real_head_count          The real number of people in that office
 * @param  int    $participating_head_count The number of participants in that office 
 * @return string                          Some HTML describing the stats totals
 */
function show_stats($who, $totals, $real_head_count, $participating_head_count){
    $grand_total = array_sum($totals);
    $reps_per_person_in_office     = (int)($grand_total / $real_head_count); 
    $reps_per_person_participating = (int)($grand_total / $participating_head_count);

    // special case: if $who is "Your", the the per person values are non-applicable
    if ($who == "Your"){
        $reps_per_person_in_office     = 'N/A';
        $reps_per_person_participating = 'N/A';
    }

    $info  = "<p>$who Totals --  " . $totals['situps'] . 
             ", " . $totals['pushups'] . 
             ", " . $totals['pullups'] . "<br />" . 
             "Total: $grand_total <br />" . 
             "Reps per person in office: " . $reps_per_person_in_office . "<br />" .
             "Reps per person participating: " . $reps_per_person_participating . "<br />";
    
    return $info;
}
