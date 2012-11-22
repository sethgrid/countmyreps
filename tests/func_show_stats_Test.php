<?php
require_once ('includes/func_show_stats.php');

class ShowStatsTest extends PHPUnit_Framework_TestCase{
    /**
     * test that the user's stats display shows correct totals
     */
    public function testWhoIsYour(){
        $stats_output    = show_stats($who = "Your", $totals = array('situps'=>10, 'pushups'=>20, 'pullups'=>30), $real_head_count=5, $participating_head_count=2);
        $expected_output = "<p>Your Totals --  10, 20, 30<br />Total: 60 <br />Reps per person in office: N/A<br />Reps per person participating: N/A<br />";
        
        $this->assertEquals($expected_output, $stats_output, "For 'Your' totals, correct totals display and N/A is used for per person counts");
    }

    /**
     * test that California's stats display shows correct totals
     */
    public function testWhoIsCalifornia(){
        $stats_output    = show_stats($who = "California", $totals = array('situps'=>10, 'pushups'=>20, 'pullups'=>30), $real_head_count=5, $participating_head_count=2);
        $expected_output = "<p>California Totals --  10, 20, 30<br />Total: 60 <br />Reps per person in office: 12<br />Reps per person participating: 30<br />";
        
        $this->assertEquals($expected_output, $stats_output, "For 'California' totals, correct totals display");
    }

    /**
     * test that Boulder's stats display shows correct totals
     */
    public function testWhoIsBoulder(){
        $stats_output    = show_stats($who = "Boulder", $totals = array('situps'=>10, 'pushups'=>20, 'pullups'=>30), $real_head_count=5, $participating_head_count=2);
        $expected_output = "<p>Boulder Totals --  10, 20, 30<br />Total: 60 <br />Reps per person in office: 12<br />Reps per person participating: 30<br />";
        
        $this->assertEquals($expected_output, $stats_output, "For 'Your' totals, correct totals display and N/A is used for per person counts");
    }

    /**
     * test that Denver's stats display shows correct totals
     */
    public function testWhoIsDenver(){
        $stats_output    = show_stats($who = "Denver", $totals = array('situps'=>10, 'pushups'=>20, 'pullups'=>30), $real_head_count=5, $participating_head_count=2);
        $expected_output = "<p>Denver Totals --  10, 20, 30<br />Total: 60 <br />Reps per person in office: 12<br />Reps per person participating: 30<br />";
        
        $this->assertEquals($expected_output, $stats_output, "For 'Denver' totals, correct totals display");
    }
}
