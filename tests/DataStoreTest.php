<?php
// modify the include path so that the required function that we are testing can load its dependencies
set_include_path("/home/seth/projects/countmyreps:/home/seth/projects/countmyreps/includes:" . get_include_path());

require_once("includes/DataStore.php");

class DataStoreTest extends PHPUnit_Extensions_Database_TestCase
{
    protected function getConnection()
    {
        $this->pdo = new PDO('mysql:host=localhost;dbname=testdb', 'root', '');
        return $this->createDefaultDBConnection($this->pdo, 'testdb');
    }

    protected function getDataSet()
    {
        return $this->createFlatXMLDataSet('tests/default_dataset.xml');
    }

    public function testUserExists()
    {
        // connect to the datastore
        $DataStore = new DataStore($this->pdo);

        // check that a user from the default dataset is preloaded
        $this->assertTrue((bool)$DataStore->user_exists('user1@example.com'));
    }
    
    public function testCreateUser()
    {
        // connect to the datastore
        $DataStore = new DataStore($this->pdo);

        // check that a non existant user can be added to the data store
        $this->assertFalse((bool)$DataStore->user_exists('non-existant-user'));
        $this->assertTrue($DataStore->create_user('non-existant-user'));
        $this->assertTrue((bool)$DataStore->user_exists('non-existant-user'));
    }

    /**
     * count returns the head count of users in that office
     */
    public function testGetCountByOffice()
    {
        // connect to the datastore
        $DataStore = new DataStore($this->pdo);

        // there are two california users
        $this->assertEquals(2, $DataStore->get_count_by_office('california')); 
    }

    /**
     * get the data for a given user, insert some reps, do it again. 
     */
    public function testGetAllRecordsByUser()
    {
        // connect to the datastore
        $DataStore = new DataStore($this->pdo);

        // go by user id
        $result_before   = $DataStore->get_all_records_by_user(1);
        $expected_before = array(
                            '2012-11-03' => array(
                                'burpees'  => 75
                            ),
                            '2012-11-04' => array(
                                'burpees'  => 105 
                            )
                        );

        $DataStore->add_reps('user1@example.com', array('burpees'=>15));
        
        $result_after   = $DataStore->get_all_records_by_user(1);
        $expected_after = array(
                            '2012-11-03' => array(
                                'burpees'  => 75,
                            ),
                            '2012-11-04' => array(
                                'burpees'  => 105,
                            ),
                            // may be brittle, second may not match in small percent of test runs
                            // possible fix would be to find if similar key exists in result and, 
                            // if so, rename key to a preset value
                            date("Y-m-d H:i:s") => array(
                                'burpees'  => 15,
                            )
                        );

        $this->assertEquals($expected_before, $result_before); 
        $this->assertEquals($expected_after,  $result_after); 
    }

    public function testGetAllRecordsByOffice()
    {
        // connect to the datastore
        $DataStore = new DataStore($this->pdo);

        $result    = $DataStore->get_all_records_by_office('california');
        $expected  = array(
                            '2012-11-03' => array(
                                'burpees'  => 75,
                            ),
                            // the 70s come from the totals of user 1 and 2 from california
                            '2012-11-04' => array(
                                'burpees'  => 210,
                            )
                     );

        $this->assertEquals($expected, $result, "Records by office are totalled correctly");
    }
}
