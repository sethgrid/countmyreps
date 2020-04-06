package countmyreps

import (
	"database/sql"
	"fmt"
	"os"
)

var exercises []Exercise
var teams []string

func init() {
	// seed the exercises that will be in the database
	// note, seeding of the db only happens if no db is present on start up.
	exercises = []Exercise{
		{
			Name:      "Push Ups",
			ValueType: "Reps",
		},
		{
			Name:      "Sit Ups",
			ValueType: "Reps",
		},
		{
			Name:      "Squats",
			ValueType: "Reps",
		},
		{
			Name:      "Pull Ups",
			ValueType: "Reps",
		},
		{
			Name:      "Burpees",
			ValueType: "Reps",
		},
		{
			Name:      "Running",
			ValueType: "Meters",
		},
	}

	// sed the teams that will be in the database
	teams = []string{
		// Offices
		"Atlanta",
		"Berlin",
		"Bogotá",
		"Denver",
		"Dublin (Block)",
		"Dublin (Wall)",
		"Hong Kong",
		"Irvine",
		"Kesklinna",
		"London",
		"Madrid",
		"Malmö",
		"Mountain View",
		"München",
		"New York",
		"Paris",
		"Praha",
		"Pyrmont",
		"Redwood City",
		"Remoties",
		"San Fransisco (Beale)",
		"San Fransisco (Spear)",
		"Singapore",
		"São Paulo",
		"Tokyo",
		"Washington DC",

		// Basic Departments
		"Engineering",
		"Finance",
		"Go-to-Market",
		"Legal",
		"Product",
	}
}

type Exercises struct {
	Collection []Exercise `json:"Exercises"`
}

type Exercise struct {
	ID        int
	Name      string
	ValueType string
}

func (s *Server) getExercises() (*Exercises, error) {
	q := "SELECT id, name, value_type FROM exercises"
	rows, err := s.DB.Query(q)
	if err != nil {
		return nil, fmt.Errorf("unable to getExercises: %w", err)
	}

	exs := &Exercises{Collection: make([]Exercise, 0)}

	for rows.Next() {
		var id int
		var name, valueType string
		err := rows.Scan(&id, &name, &valueType)
		if err != nil {
			return nil, fmt.Errorf("unable to scan getExercises: %w", err)
		}
		exs.Collection = append(exs.Collection, Exercise{ID: id, Name: name, ValueType: valueType})
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("unexpected error after scanning getExercises: %w", err)
	}

	return exs, nil
}

func (s *Server) getOrCreateUser(email string) (int, error) {
	q := "select id from users where email = ?;"
	row := s.DB.QueryRow(q, email)

	var id int
	err := row.Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("unable to scan users: %w", err)
	}

	if id != 0 {
		return id, nil
	}

	// no id returned; time to create the user

	stmt := "insert into users (email) values (?);"
	res, err := s.DB.Exec(stmt, email)
	if err != nil {
		return 0, fmt.Errorf("unable to insert into users: %w", err)
	}
	newID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("unable to get id from new insert into users: %w", err)
	}

	// could overflow in a 32 bit system. Very unlikely in our case as we are limited to twilio.com addrs :p
	return int(newID), nil
}

func (s *Server) InitDB() error {
	_, err := os.Stat(s.conf.DBPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot stat file - %w", err)
	}

	if os.IsNotExist(err) {
		if err := s.seedDB(); err != nil {
			return fmt.Errorf("unable to seed db - %w", err)
		}
	}

	db, err := sql.Open("sqlite3", s.conf.DBPath)
	if err != nil {
		return fmt.Errorf("cannot open db - %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("cannot ping db - %w", err)
	}

	s.DB = db

	return nil
}

func (s *Server) seedDB() error {

	db, err := sql.Open("sqlite3", s.conf.DBPath)
	if err != nil {
		return fmt.Errorf("unable to open db at %q - %w", s.conf.DBPath, err)
	}

	createTables := []string{
		"create table exercises (id integer not null primary key autoincrement, name text, value_type text);",
		"create table users (id integer not null primary key autoincrement, email text, created_on timestamp default current_timestamp);",
		"create table reps (id integer not null primary key autoincrement, exercise_id integer, user_id integer, count integer, created_on timestamp default current_timestamp);",
		"create table teams (id integer not null primary key autoincrement, name text, created_by_user_id integer);",
		"create table user_teams (id integer not null primary key autoincrement, team_id integer, user_id integer);",
	}

	for _, stmt := range createTables {
		_, err = db.Exec(stmt)
		if err != nil {
			return fmt.Errorf("unable to create table '%s' - %w", stmt, err)
		}
	}

	// seed with exercises
	stmt := "insert into exercises (name, value_type) values(?, ?);"
	for _, exercise := range exercises {
		_, err := db.Exec(stmt, exercise.Name, exercise.ValueType)
		if err != nil {
			return fmt.Errorf("unable to insert %#v into exercises - %w", exercise, err)
		}
	}

	// seed with teams
	stmt = "insert into teams (name, created_by_user_id) values (?, -1)"
	for _, teamName := range teams {
		_, err := db.Exec(stmt, teamName)
		if err != nil {
			return fmt.Errorf("unable to insert %s into teams - %w", teamName, err)
		}
	}

	return nil
}
