package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/facebookgo/flagenv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// ViewTemplate displays /view
var ViewTemplate *template.Template

// IndexTemplate displays the index/root
var IndexTemplate *template.Template

// StartDate is the earliest date we will query in the db
var StartDate time.Time

// EndDate is the latest date we will query in the db
var EndDate time.Time

// Offices is all the valid Offices
var Offices []string

// AppName is the app name
var AppName = "countmyreps"

// Version is the semver
var Version = "3.1.3"

func init() {
	var err error
	rand.Seed(time.Now().UnixNano())

	// funcMap contains the functions available to the view template
	funcMap := template.FuncMap{
		// totals sums all the exercises in []RepData
		"totals": func(d []RepData) int {
			return totalReps(d)
		},
		// allow easy converting of strings to JS string (turns freqData{{ OfficeName}}: freqData"OC" -> freqDataOC in JS)
		"js": func(s string) template.JS {
			return template.JS(s)
		},
		// d3ChartData correctly formats []RepData to the JS format so data can display
		"d3ChartData": func(d []RepData) template.JS {
			parts := make([]string, len(d))
			for i, data := range d {
				parts[i] = fmt.Sprintf("{State:'%s',freq:{pull_up:%d, sit_up:%d, push_up: %d, squat:%d}}",
					data.Date,
					data.ExerciseCounts[PullUps],
					data.ExerciseCounts[SitUps],
					data.ExerciseCounts[PushUps],
					data.ExerciseCounts[Squats],
				)
			}
			return template.JS(strings.Join(parts, ",\n"))
		},
		// d3ChartDataForOffice is a helper method to avoid complexities with nesting ranges in the template
		"d3ChartDataForOffice": func(officeName string, reps map[string][]RepData) template.JS {
			// TODO: DRY up with ^^
			parts := make([]string, len(reps[officeName]))
			for i, data := range reps[officeName] {
				parts[i] = fmt.Sprintf("{State:'%s',freq:{pull_up:%d, sit_up:%d, push_up: %d, squat:%d}}",
					data.Date,
					data.ExerciseCounts[PullUps],
					data.ExerciseCounts[SitUps],
					data.ExerciseCounts[PushUps],
					data.ExerciseCounts[Squats],
				)
			}
			return template.JS(strings.Join(parts, ",\n"))
		},
	}

	// parse tempaltes in init so we don't have to parse them on each request
	// pro: single parsing
	// con: have to restart the process to load file changes
	ViewTemplate, err = template.New("view.html").Funcs(funcMap).ParseFiles(filepath.Join("go_templates", "view.html"))
	if err != nil {
		log.Fatalln(err)
	}

	IndexTemplate, err = template.New("index.html").Funcs(funcMap).ParseFiles(filepath.Join("go_templates", "index.html"))
	if err != nil {
		log.Fatalln(err)
	}

}

// We expect the database to have these exact values
// TODO: it would be interesting to figure out how to have this dynamic (display too) and based off the email they send to
const (
	PullUps = "Pull Ups"
	SitUps  = "Sit Ups"
	PushUps = "Push Ups"
	Squats  = "Squats"
)

var (
	OldEmail = "pullups-pushups-airsquats-situps@countmyreps.com"
	NewEmail = "pullups-pushups-squats-situps@countmyreps.com"
)

// Debug turns on more verbose logging
var Debug bool

// EmailSender allows us to swap out how we send the email (specifically, SendGrid vs Fake/Test)
var EmailSender Emailer

func main() {
	var err error

	// flag vars
	var port int
	var mysqlHost, mysqlPort, mysqlUser, mysqlPass, mysqlDBname string
	var start, end string

	// defaults for start and end vars
	startDefault := fmt.Sprintf("%d-11-01", time.Now().Year())
	endDefault := fmt.Sprintf("%d-11-30", time.Now().Year())

	// get flags
	flag.IntVar(&port, "port", 9126, "port to run site")
	flag.StringVar(&start, "start-date", startDefault, "the start date to when querying the db")
	flag.StringVar(&end, "end-date", endDefault, "the end date to when querying the db")
	flag.StringVar(&mysqlHost, "mysql-host", "localhost", "mysql host")
	flag.StringVar(&mysqlPort, "mysql-port", "3306", "mysql port")
	flag.StringVar(&mysqlUser, "mysql-user", "root", "mysql root")
	flag.StringVar(&mysqlPass, "mysql-pass", "", "mysql pass")
	flag.StringVar(&mysqlDBname, "mysql-dbname", "countmyreps", "mysql dbname")
	flag.StringVar(&OldEmail, "old-email", "pullups-pushups-airsquats-situps@countmyreps.com", "email to receive from")
	flag.StringVar(&NewEmail, "new-email", "pullups-pushups-squats-situps@countmyreps.com", "email to receive from")
	flag.BoolVar(&Debug, "debug", false, "set flag for verbose logging")

	flagenv.Parse()
	flag.Parse()

	// validate flags
	StartDate, err = time.Parse("2006-01-02", start)
	if err != nil {
		log.Fatal("err parsing date", err)
	}
	EndDate, err = time.Parse("2006-01-02", end)
	if err != nil {
		log.Fatal("err parsing date", err)
	}

	db := SetupDB(mysqlUser, mysqlPass, mysqlHost, mysqlPort, mysqlDBname)

	log.Printf("starting on :%d", port)
	s := NewServer(db, port, SendGridEmailer{})

	if err := s.Serve(); err != nil {
		log.Println("Unexpected error serving: ", err.Error())
	}
}

// SetupDB initialized the DB conn and grabs initial data needed for the app (ie, Offices)
func SetupDB(mysqlUser, mysqlPass, mysqlHost, mysqlPort, mysqlDBname string) *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", mysqlUser, mysqlPass, mysqlHost, mysqlPort, mysqlDBname))
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(1 * time.Minute)

	err = populateOfficesVar(db)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// Server contains the settings needed to run the server
type Server struct {
	Port int
	Mux  *mux.Router
	DB   *sql.DB

	dbname string
	close  chan struct{}
}

// NewServer creates a new server running against the give db
func NewServer(db *sql.DB, port int, emailer Emailer) *Server {
	s := &Server{}
	s.Port = port
	s.DB = db
	s.close = make(chan struct{})
	EmailSender = emailer // TODO: should this be on the server? How will that pass down?

	r := mux.NewRouter()
	r.HandleFunc("/", s.IndexHandler)
	r.HandleFunc("/view", s.ViewHandler)
	r.HandleFunc("/json", s.JSONHandler)
	r.HandleFunc("/healthcheck", s.HealthcheckHandler)
	r.HandleFunc("/parseapi/index.php", s.ParseHandler)                                // backwards compatibility
	r.PathPrefix("/").Handler(http.StripPrefix("", http.FileServer(http.Dir("web/")))) // mux specific workaround for fileserver; todo: use separate mux to avoid filtering these endpoints from logs?

	r.Handle("/", mwPanic(mwLog(r)))

	s.Mux = r
	return s
}

// Serve blocks and starts a server
func (s *Server) Serve() error {
	errc := make(chan error)
	if s.Port == 0 {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			errc <- err
		}
		s.Port = l.Addr().(*net.TCPAddr).Port
		errc <- http.Serve(l, s.Mux)
	} else {
		errc <- http.ListenAndServe(fmt.Sprintf(":%d", s.Port), s.Mux)
	}
	for {
		select {
		case <-s.close:
			return nil
		case err := <-errc:
			return err
		}
	}
}

// Close terminates the server
func (s *Server) Close() error {
	close(s.close)
	return nil
}

// errorHandler is a helper method to log and display errors. When invoked from a parent handler, the parent should then return
func errorHandler(w http.ResponseWriter, r *http.Request, code int, message string, err error) {
	logError(r, err, message)
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf("%v - %s", http.StatusText(code), message)))
}

func getMessageIDFromHeader(s string) string {
	r := regexp.MustCompile("\r?\n")
	for _, line := range r.Split(s, -1) {
		if strings.Contains(line, "Message-ID") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				return ""
			}

			return strings.TrimSpace(parts[1])
		}
	}

	return ""
}

// ParseHandler handles SendGrid's inbound parse api
func (s *Server) ParseHandler(w http.ResponseWriter, r *http.Request) {
	// NOTE: SendGrid's Inbound Parse API requires a 200 level response always, even on error, otherwise it will retry

	// errMsg is parsed later to determine if we should send a success or error email
	var errMsg string
	var err error

	to := r.PostFormValue("to")
	from := r.PostFormValue("from")
	subject := r.PostFormValue("subject")
	headers := r.PostFormValue("headers")
	replyMessageID := getMessageIDFromHeader(headers)

	logEvent(r, "parseapi", fmt.Sprintf("To: %s, From: %s, Subject: %s", to, from, subject))

	defer func() {
		var mailType string
		if errMsg != "" {
			mailType = "error - " + errMsg
			// we don't want to send out a bunch of responses to spam hitting the server
			// only send a response if the subject looked vaguely correct or the sender was from sendgrid.
			parts := strings.Split(subject, ",")
			if strings.Contains(from, "@sendgrid.com") || len(parts) == 4 {
				err = s.SendErrorEmail(from, to, subject, errMsg, replyMessageID)
			}
		} else {
			mailType = "success"
			err = s.SendSuccessEmail(from, subject, replyMessageID)
		}
		if err != nil {
			logError(r, err, "unable to send response email: "+mailType)
		}
	}()

	if to == "" || from == "" || subject == "" {
		logEvent(r, "bad_parse", "unable to determine to or from or subject")
		errMsg = fmt.Sprintf(ErrUnexpectedFmt, fmt.Sprintf("Missing to, from, or subject: %q, %q, %q", to, from, subject))
		return
	}

	from = extractEmailAddr(from)

	if !strings.Contains(from, "@sendgrid.com") {
		logEvent(r, "bad_parse", fmt.Sprintf("sender not from sendgrid - %s", from))
		errMsg = fmt.Sprintf(ErrToAddrFmt, from)
		return
	}

	if !(extractEmailAddr(to) == NewEmail || extractEmailAddr(to) == OldEmail) {
		logEvent(r, "bad_parse", fmt.Sprintf("recipient not valid countmyreps address: %s", to))
		errMsg = fmt.Sprintf(ErrToAddrFmt, to)
		return
	}

	userID, err := getOrCreateUserID(s.DB, from)
	if err != nil {
		logError(r, err, "unable to create/get user")
		errMsg = fmt.Sprintf(ErrUnexpectedFmt, "unable to create and/or get user")
		return
	}

	reps := strings.Split(subject, ",")
	if len(reps) == 4 {
		for i, rep := range reps {
			count, err := strconv.Atoi(strings.TrimSpace(rep))
			if err != nil {
				logError(r, err, fmt.Sprintf("unable to convert %s to int", rep))
				errMsg = fmt.Sprintf(ErrSubjectFmt, subject)
				return
			}
			// protect against tricky people who spoof negative reps to other folks
			if count < 0 {
				count = -1 * count
			}
			var exercise string
			switch i {
			case 0:
				exercise = PullUps
			case 1:
				exercise = PushUps
			case 2:
				exercise = Squats
			case 3:
				exercise = SitUps
			default:
				exercise = "unknown"
			}
			// TODO: move out of loop and use VALUES (), (), (), () and move to db.go
			_, err = s.DB.Exec("INSERT INTO reps (exercise, count, user_id) VALUES (?, ?, ?)", exercise, count, userID)
			if err != nil {
				logError(r, err, "unable to insert rep")
				errMsg = fmt.Sprintf(ErrUnexpectedFmt, "unable to insert into the database")
				return
			}
		}
	} else if inListCaseInsenitive(subject, Offices) {
		// TODO: remove offices; just use teams
		office := formattedOffice(subject)
		// todo: move to db.go
		_, err = s.DB.Exec("UPDATE user SET office=(SELECT id FROM office where name=?) WHERE id=? LIMIT 1", office, userID)
		if err != nil {
			logError(r, err, "unable to update user's office")
			errMsg = fmt.Sprintf(ErrUnexpectedFmt, "unable to update office relationship in the database")
			return
		}
	} else if strings.Contains(strings.ToLower(subject), "team add:") {
		parts := strings.Split(subject, ":")
		if len(parts) < 2 {
			logError(r, err, "enexpected error splitting subject for team add")
			errMsg = fmt.Sprintf(ErrUnexpectedFmt, "unable to add to user teams")
			return
		}
		err = addTeam(s.DB, sanitizeTeamName(parts[1]), userID)
		if err != nil {
			logError(r, err, "unable to add to user teams")
			errMsg = fmt.Sprintf(ErrUnexpectedFmt, "unable to add to user teams")
			return
		}
	} else if strings.Contains(strings.ToLower(subject), "team remove:") {
		parts := strings.Split(subject, ":")
		if len(parts) < 2 {
			logError(r, err, "enexpected error splitting subject for team remove")
			errMsg = fmt.Sprintf(ErrUnexpectedFmt, "unable to add to user teams")
			return
		}
		err = removeTeam(s.DB, sanitizeTeamName(parts[1]), userID)
		if err != nil {
			logError(r, err, "unable to remove from user teams")
			errMsg = fmt.Sprintf(ErrUnexpectedFmt, "unable to remove from user teams")
			return
		}
	} else {
		logEvent(r, "bad_parse", fmt.Sprintf("bad subject: %s", subject))
		errMsg = fmt.Sprintf(ErrSubjectFmt, subject)
		return
	}
}

func sanitizeTeamName(teamName string) string {
	var rns []rune
	for _, r := range teamName {
		if strings.ContainsAny(string(r), "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_") {
			rns = append(rns, r)
		}
	}
	return string(rns)
}

// ViewData is the data needed to populate the view.html template
type ViewData struct {
	UserEmail  string
	UserOffice string
	UserTeams  []string
	TodaysReps []RepData
	UserReps   []RepData
	TeamReps   map[string][]RepData
	TeamStats  map[string]Stats
}

// RepData is a single entry (or aggregate for a day)
type RepData struct {
	Date           string
	ExerciseCounts map[string]int
}

// Stats hold per-office stats
type Stats struct {
	RepsPerPerson                    int
	RepsPerPersonParticipating       int
	RepsPerPersonPerDay              int
	RepsPerPersonParticipatingPerDay int
	PercentParticipating             int
	TotalReps                        int
	HeadCount                        int
}

// ViewHandler handles /view (all the graphs, data, etc)
func (s *Server) ViewHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		errorHandler(w, r, http.StatusBadRequest, "you must provide an email query parameter", nil)
		return
	}

	data := s.getViewData(email)

	err := ViewTemplate.Execute(w, data)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to execute %s template", "view.html"), err)
		return
	}
}

// JSONHandler displays the JSON payload needed to build a client based js page
func (s *Server) JSONHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		errorHandler(w, r, http.StatusBadRequest, "you must provide an email query parameter", nil)
		return
	}

	data := s.getViewData(email)

	w.Header().Set("content-type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, "unable to encode json", err)
	}
}

func (s *Server) getViewData(email string) ViewData {
	// TODO: clean up this hack by migrating user offices to user be one of their teams
	officeAndTeamReps := getTeamReps(s.DB)
	for k, v := range getOfficeReps(s.DB) {
		officeAndTeamReps[k] = v
	}
	officeAndTeamStats := getTeamStats(s.DB)
	for k, v := range getOfficeStats(s.DB) {
		officeAndTeamStats[k] = v
	}

	data := ViewData{
		UserEmail:  email,
		TodaysReps: getTodaysReps(s.DB, email),
		UserOffice: getUserOffice(s.DB, email),
		UserTeams:  getUserTeams(s.DB, email),
		TeamReps:   officeAndTeamReps,
		TeamStats:  officeAndTeamStats,
		UserReps:   getUserReps(s.DB, email),
	}
	return data
}

// HeathcheckHandler verifies dependencies and reports if they are not in a good state
func (s *Server) HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	_, err := s.DB.Exec("SELECT 1")
	if err != nil {
		logError(r, err, "healthcheck failed to query db")
		w.Write([]byte("database issues\n"))
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.Write([]byte("database ok\n"))
	}
}

// IndexHandler handles the root/index
func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	err := IndexTemplate.Execute(w, nil)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to execute %s template", "index.html"), err)
		return
	}
}
