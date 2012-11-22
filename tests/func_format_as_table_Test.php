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
        $expected_table_headers = "<th>Date/Time</th><th>Sit-ups</th><th>Push-ups</th><th>Pull-ups</th>";
        $expected_table_row_1   = "<tr><td>2012-11-03</td><td>15</td><td>10</td><td>5</td></tr>";
        $expected_table_row_2   = "<tr><td>2012-11-04</td><td>15</td><td>10</td><td>5</td></tr>";
        $expected_table_row_3   = "<tr><td>2012-11-05</td><td>15</td><td>10</td><td>5</td></tr>";
        
        $this->assertRegExp("~$expected_table_headers~", $result, "table headers match");
        $this->assertRegExp("~$expected_table_row_1~",   $result, "table row match");
        $this->assertRegExp("~$expected_table_row_2~",   $result, "table row match");
        $this->assertRegExp("~$expected_table_row_3~",   $result, "table row match");
    }
}
