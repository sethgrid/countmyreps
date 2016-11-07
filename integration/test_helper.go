package integration

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SQLTables is the file in the setup dir
const SQLTables = "create_db_v2.sql"

// Debug will make logging verbose
var Debug bool

func SetupDB(dsn string, dbname string, overwrite bool) *sql.DB {
	var err error

	// initial connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		debugFatalf("unable to connect to the db - %v", err)
	}
	if err := db.Ping(); err != nil {
		debugFatalf("unable to ping the db - %v", err)
	}

	// create the database
	debugf("attempting to create %s", dbname)
	_, err = db.Exec(fmt.Sprintf("create database %s", dbname))
	if err != nil && strings.Contains(err.Error(), "database exists") {
		if !overwrite {
			debugFatalf("unable to proceed; database %s exists. Please use the -overwrite-database flag to continue.", dbname)
		}
		debugf("dropping %s", dbname)
		_, err = db.Exec(fmt.Sprintf("drop database %s", dbname))
		if err != nil {
			debugFatalf("unable to drop %s: %v", dbname, err)
		}
		debugf("recreated database %s", dbname)
		_, err = db.Exec(fmt.Sprintf("create database %s", dbname))
		if err != nil {
			debugFatalf("unable to create database: %v", err)
		}

	} else if err != nil {
		debugFatalf("unable to create db %s: %v", dbname, err)
	}

	_, err = db.Exec("use " + dbname)
	if err != nil {
		debugFatalf("unable to select database: %v", err)
	}

	// create tables
	debugln("ready to create tables")

	dir, err := os.Getwd()
	if err != nil {
		debugFatal(err)
	}

	setupData, err := ioutil.ReadFile(filepath.Join(dir, "..", "setup", SQLTables))
	if err != nil {
		setupData, err = ioutil.ReadFile(filepath.Join(dir, "setup", SQLTables))
		if err != nil {
			debugf("unable to read setup/%s: %v", SQLTables, err)
		}
	}

	createTables := strings.Split(string(setupData), ";")
	for _, createTable := range createTables {
		createTable = strings.TrimSpace(createTable)
		if createTable != "" {
			_, err = db.Exec(createTable)
			if err != nil {
				debugFatalf("unable to create tables: %v\n%s", err, createTable)
			}
		}
	}

	debugln("tables set up")

	return db
}

// TearDownDB drops the given database
func TearDownDB(db *sql.DB, dbname string) {
	_, err := db.Exec("drop database " + dbname)
	if err != nil {
		debugFatalf("unable to drop %s: %v", dbname, err)
	}
}

// Seed populates sample data into the DB
func Seed(db *sql.DB, monthStart string, monthEnd string, today string) error {
	OCHeadCount := 4
	DenverHeadCount := 8
	start, err := time.Parse("2006-01-02", monthStart)
	if err != nil {
		return err
	}
	// end, err := time.Parse("2006-01-02", monthEnd)
	// if err != nil {
	// 	return err
	// }
	now, err := time.Parse("2006-01-02", today)
	if err != nil {
		return err
	}

	// create offices
	debugln("inserting offices")
	_, err = db.Exec("INSERT INTO office (name, head_count) VALUES ('', 1), ('OC', ?), ('Denver', ?)", OCHeadCount, DenverHeadCount)
	if err != nil {
		return err
	}

	// create users
	var users []User
	debugln("inserting users")
	for i := 1; i <= OCHeadCount; i++ {
		user := fmt.Sprintf("oc_%d@sendgrid.com", i)
		res, err := db.Exec("INSERT INTO user SET email=?, office=(SELECT id FROM office WHERE name='OC' LIMIT 1)", user)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		users = append(users, User{id: id, email: user})
	}

	for i := 1; i <= DenverHeadCount; i++ {
		user := fmt.Sprintf("denver_%d@sendgrid.com", i)
		res, err := db.Exec("INSERT INTO user SET email=?, office=(SELECT id FROM office WHERE name='Denver' LIMIT 1)", user)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		users = append(users, User{id: id, email: user})
	}

	// create reps
	debugln("inserting rep data")
	var randSeed int64 = 655321 // always get the same test data
	r := rand.New(rand.NewSource(randSeed))
	for itr := start; !itr.After(now); itr = itr.Add(time.Hour * 24) {
		thisDay := itr.Format("2006-01-02")
		for _, user := range users {
			if user.id%2 == 0 {
				continue
			}
			_, err := db.Exec("INSERT INTO reps (exercise, count, user_id, created_at) VALUES ('Pull Ups', ?, ?, ?), ('Push Ups', ?, ?, ?), ('Sit Ups', ?, ?, ?), ('Squats', ?, ?, ?)",
				int(user.id)*r.Intn(5), user.id, thisDay,
				int(user.id)*r.Intn(10), user.id, thisDay,
				int(user.id)*r.Intn(15), user.id, thisDay,
				int(user.id)*r.Intn(20), user.id, thisDay,
			)
			if err != nil {
				return err
			}
		}
	}

	debugln("database seeded")
	return nil
}

type User struct {
	id    int64
	email string
}

func debugln(v ...interface{}) {
	if Debug {
		log.Println(v...)
	}
}

func debugf(format string, v ...interface{}) {
	if Debug {
		log.Printf(format, v...)
	}
}

func debugFatal(v ...interface{}) {
	if Debug {
		log.Fatal(v...)
	}
}

func debugFatalf(format string, v ...interface{}) {
	if Debug {
		log.Fatalf(format, v...)
	}
}
