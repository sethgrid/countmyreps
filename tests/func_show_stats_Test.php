<?php
require_once ('includes/func_show_stats.php');

class ShowStatsTest extends PHPUnit_Framework_TestCase{
    /**
     * test that the user's stats display shows correct totals
     */
    public function testWhoIsYour(){
        $stats_output    = show_stats($who = "Your", $totals = array('burpees'=>30), $real_head_count=5, $participating_head_count=2);
        $expected_output = "<p>Your Total: 30<br />Reps per person in office: N/A<br />Reps per person participating: N/A<br />Percent participating: N/A%<br />";
        
        $this->assertEquals($expected_output, $stats_output, "For 'Your' totals, correct totals display and N/A is used for per person counts");
    }

    /**
     * test that California's stats display shows correct totals
     */
    public function testWhoIsCalifornia(){
        $stats_output    = show_stats($who = "California", $totals = array('burpees'=>30), $real_head_count=5, $participating_head_count=2);
        $expected_output = "<p>California Total: 30<br />Reps per person in office: 6<br />Reps per person participating: 15<br />Percent participating: 40%<br />";
        
        $this->assertEquals($expected_output, $stats_output, "For 'California' totals, correct totals display");
    }
}
