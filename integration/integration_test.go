package integration

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/facebookgo/flagenv"
	_ "github.com/go-sql-driver/mysql"
)

const SQLTables = "create_db_v2.sql"

// Port can be overridden with flags
var Port int

func TestMain(m *testing.M) {
	// flag vars
	var mysqlHost, mysqlPort, mysqlUser, mysqlPass, mysqlDBname string
	var start, end string
	var overwriteDB, noTearDown bool

	// defaults for start and end vars
	monthStart := "2016-11-01"
	monthEnd := "2016-11-30"
	today := "2016-11-15"

	flag.IntVar(&Port, "port", 9126, "port of the running countmyreps instance")
	flag.StringVar(&start, "start-date", monthStart, "the start date to when querying the db")
	flag.StringVar(&end, "end-date", monthEnd, "the end date to when querying the db")
	flag.StringVar(&today, "today-date", today, "the date that we set the world to be for the tests")
	flag.StringVar(&mysqlHost, "mysql-host", "localhost", "mysql host")
	flag.StringVar(&mysqlPort, "mysql-port", "3306", "mysql port")
	flag.StringVar(&mysqlUser, "mysql-user", "root", "mysql root")
	flag.StringVar(&mysqlPass, "mysql-pass", "", "mysql pass")
	flag.StringVar(&mysqlDBname, "mysql-dbname", "countmyreps_test", "mysql dbname")
	flag.BoolVar(&overwriteDB, "overwrite-database", false, "allow overwriting database")
	flag.BoolVar(&noTearDown, "no-tear-down", false, "set to keep the data after the test run")

	flagenv.Parse()
	flag.Parse()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", mysqlUser, mysqlPass, mysqlHost, mysqlPort)

	db := setupDB(dsn, mysqlDBname, overwriteDB)
	err := seed(db, monthStart, monthEnd, today)
	if err != nil {
		log.Fatalf("error seeding data: %v", err)
	}
	if !noTearDown {
		defer tearDownDB(db, mysqlDBname)
	}
	// TODO - start countmyreps
	os.Exit(m.Run())
}

func TestStats(t *testing.T) {
	resp, err := http.Get("http://locahost:%d/view?email=oc_1%40sendgrid.com", Port)
	if err != nil {
		t.Fatalf("unable to get countmyreps data: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Errorf("got %d, want %d: %s", got, want, body)
	}

	// TODO - parse dom? how best to figure out if data is good? Wait until we have a json endpiont?
}
