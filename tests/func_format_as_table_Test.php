<?php
require_once ('includes/func_format_as_table.php');

class FormatAsTableTest extends PHPUnit_Framework_TestCase{
    /**
     * test that the user's data shows as a table
     */
    public function testCleanRun(){
        $data = array(
                  "2012-11-03" => array("situps"=>15, "pushups"=>10, "pullups"=>5),
                  "2012-11-04" => array("situps"=>15, "pushups"=>10, "pullups"=>5),
                  "2012-11-05" => array("situps"=>15, "pushups"=>10, "pullups"=>5),
                );

        $result = format_as_table($data);
        $expected_table_headers = "<th>Date/Time</th><th>Burpees</th>";
        
        $this->assertRegExp("~$expected_table_headers~", $result, "table headers match");
    }
}
