<?php

function get_totals($data_array){
    print_r($data_array);

    $return = Array();
    $return['situps'] = 3;
    $return['pushups'] = 2;
    $return['pullups'] = 1;

    return $return;
       
}
