package main

import (
	"bufio"
	"context"
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
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/facebookgo/flagenv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// TODO: have page layout in viewHandler require no manual updating for new offices.

// ViewTemplate displays /view
var ViewTemplate *template.Template

// IndexTemplate displays the index/root
var IndexTemplate *template.Template

// StartDate is the earliest date we will query in the db
var StartDate time.Time

// EndDate is the latest date we will query in the db
var EndDate time.Time

// DB is the mysql db instance
var DB *sql.DB

func init() {
	var err error
	rand.Seed(time.Now().UnixNano())

	// funcMap contains the functions available to the view template
	funcMap := template.FuncMap{
		// totals sums all the exercises in []RepData
		"totals": func(d []RepData) int {
			sum := 0
			for _, rd := range d {
				for _, count := range rd.ExerciseCounts {
					sum += count
				}
			}
			return sum
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
					data.ExerciseCounts[PushUps],
					data.ExerciseCounts[SitUps],
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
					data.ExerciseCounts[PushUps],
					data.ExerciseCounts[SitUps],
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
// TODO: it would be interesting to figure out how to have this dynamic (display too)
const (
	PullUps = "Pull Ups"
	SitUps  = "Sit Ups"
	PushUps = "Push Ups"
	Squats  = "Squats"
)

func main() {
	var err error

	// flag vars
	var port int
	var mysqlHost, mysqlPort, mysqlUser, mysqlPass, mysqlDBname string
	var start, end string

	// defaults for start and end vars
	thisOctoberStart := fmt.Sprintf("%d-10-01", time.Now().Year())
	thisOctoberEnd := fmt.Sprintf("%d-10-31", time.Now().Year())

	// get flags
	flag.IntVar(&port, "port", 9126, "port to run site")
	flag.StringVar(&start, "start-date", thisOctoberStart, "the start date to when querying the db")
	flag.StringVar(&end, "end-date", thisOctoberEnd, "the end date to when querying the db")
	flag.StringVar(&mysqlHost, "mysql-host", "localhost", "mysql host")
	flag.StringVar(&mysqlPort, "mysql-port", "3306", "mysql port")
	flag.StringVar(&mysqlUser, "mysql-user", "root", "mysql root")
	flag.StringVar(&mysqlPass, "mysql-pass", "", "mysql pass")
	flag.StringVar(&mysqlDBname, "mysql-dbname", "countmyreps", "mysql dbname")

	flagenv.Parse()
	flag.Parse()

	// validate flags
	StartDate, err = time.Parse("2006-01-02", start)
	if err != nil {
		log.Fatal(err)
	}
	EndDate, err = time.Parse("2006-01-02", end)
	if err != nil {
		log.Fatal(err)
	}

	// connect to the db
	DB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", mysqlUser, mysqlPass, mysqlHost, mysqlPort, mysqlDBname))
	if err != nil {
		log.Fatal(err)
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	DB.SetConnMaxLifetime(1 * time.Minute)

	// set up routes and serve
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/view", viewHandler)
	r.PathPrefix("/web/").Handler(http.StripPrefix("/web/", http.FileServer(http.Dir("web/")))) // mux specific workaround for fileserver; todo: use separate mux to avoid log filtering?

	http.Handle("/", mwPanic(mwLog(r)))

	log.Printf("starting on :%d", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Println("Unexpected error serving: ", err.Error())
	}
}

// errorHandler is a helper method to log and display errors. When invoked from a parent handler, the parent should then return
func errorHandler(w http.ResponseWriter, r *http.Request, code int, message string, err error) {
	logError(r, err, message)
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf("%v - %s", http.StatusText(code), message)))
}

// ViewData is the data needed to populate the view.html template
type ViewData struct {
	UserEmail   string
	UserOffice  string
	TodaysReps  []RepData
	UserReps    []RepData
	OfficeReps  map[string][]RepData
	OfficeStats map[string]Stats
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
	OfficeSize                       int
}

// viewHandler handles /view (all the graphs, data, etc)
func viewHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		errorHandler(w, r, http.StatusBadRequest, "you must provide an email query parameter", nil)
		return
	}
	data := ViewData{
		UserEmail:   email,
		TodaysReps:  getTodaysReps(email),
		UserOffice:  getUserOffice(email),
		OfficeReps:  getOfficeReps(),
		OfficeStats: getOfficeStats(),
		UserReps:    getUserReps(email),
	}

	err := ViewTemplate.Execute(w, data)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to execute %s template", "view.html"), err)
		return
	}
}

// indexHandler handles the root/index
func indexHandler(w http.ResponseWriter, r *http.Request) {
	err := IndexTemplate.Execute(w, nil)
	if err != nil {
		errorHandler(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to execute %s template", "index.html"), err)
		return
	}
}

func getTodaysReps(email string) []RepData {
	var rd []RepData
	q := "SELECT reps.exercise, reps.count, reps.created_at FROM reps JOIN user on reps.user_id=user.id WHERE user.email=? AND created_at >= ?"
	rows, err := DB.Query(q, email, fmt.Sprintf("%d-%d-%d", time.Now().Year(), int(time.Now().Month()), time.Now().Day()))
	if err != nil {
		logError(nil, err, "unable to get today's reps")
		return rd
	}
	for rows.Next() {
		var exercise string
		var count int
		var createdAt time.Time
		err := rows.Scan(&exercise, &count, &createdAt)
		if err != nil {
			logError(nil, err, "unable to scan today's reps")
		}
		rd = append(rd, RepData{
			Date:           createdAt.Format(time.Kitchen),
			ExerciseCounts: map[string]int{exercise: count},
		})
	}
	return rd
}

func getUserOffice(email string) string {
	var officeName string
	q := "SELECT office.name FROM user JOIN office ON user.office=office.id WHERE user.email=?"
	row := DB.QueryRow(q, email)
	err := row.Scan(&officeName)
	if err != nil {
		logError(nil, err, "unable to query for office name")
		return ""
	}
	return officeName
}

func queryPrinter(q string, args ...interface{}) string {
	qFmt := strings.Replace(q, "?", `"%v"`, -1)
	return fmt.Sprintf(qFmt, args...)
}

func getOfficeStats() map[string]Stats {
	officeStats := make(map[string]Stats)
	for _, officeName := range []string{"OC", "RWC", "Denver", "Euro"} {
		var headCount int
		var participating int
		var totalReps sql.NullInt64

		qHeadCount := "SELECT head_count FROM office WHERE office.name=?"
		row := DB.QueryRow(qHeadCount, officeName)
		err := row.Scan(&headCount)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			logError(nil, err, "unable to scan for office head count")
		}

		qParticip := "SELECT count(distinct id) from (SELECT user.id FROM reps JOIN user on reps.user_id=user.id JOIN office ON user.office=office.id WHERE office.name=? and reps.created_at > ? AND reps.created_at < ?) participating;"
		row = DB.QueryRow(qParticip, officeName, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"))
		err = row.Scan(&participating)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			logError(nil, err, "unable to scan for office participation")
			return officeStats
		}

		qTotals := "select sum(reps.count) from reps left join user on reps.user_id=user.id join office on office.id=user.office where reps.created_at > ? and reps.created_at < ? and office.name=?;"
		// log.Println(queryPrinter(qTotals, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"), officeName))
		row = DB.QueryRow(qTotals, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"), officeName)
		err = row.Scan(&totalReps)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			logError(nil, err, "unable to scan for office totals")
			return officeStats
		}

		totalDays := int(EndDate.Sub(StartDate).Hours() / float64(24))
		if totalDays <= 0 {
			totalDays = 1 // avoid divide by zero
		}

		if headCount == 0 {
			headCount = 1 // avoid divide by zero
		}
		stats := Stats{}
		stats.OfficeSize = headCount
		stats.TotalReps = int(totalReps.Int64)
		stats.PercentParticipating = participating * 100 / headCount
		stats.RepsPerPerson = int(totalReps.Int64) / headCount

		if participating == 0 {
			participating = 1 // avoid divide by zero
		}
		stats.RepsPerPersonParticipating = int(totalReps.Int64) / participating
		stats.RepsPerPersonParticipatingPerDay = int(totalReps.Int64) / participating / totalDays
		stats.RepsPerPersonPerDay = int(totalReps.Int64) / headCount / totalDays

		officeStats[officeName] = stats
	}
	return officeStats
}

func getOfficeReps() map[string][]RepData {
	officeReps := make(map[string][]RepData)
	// TODO: DRY it up getUserReps
	for _, officeName := range []string{"OC", "RWC", "Denver", "Euro"} {
		q := "SELECT reps.exercise, reps.count, reps.created_at FROM reps JOIN user on reps.user_id=user.id WHERE user.id in (SELECT user.id FROM user JOIN office on user.office=office.id WHERE office.name=?) AND reps.created_at > ? AND reps.created_at < ?"
		rows, err := DB.Query(q, officeName, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"))
		if err != nil {
			logError(nil, err, "unable to query for user's reps")
			return nil
		}
		repDatas := initRepData()
		for rows.Next() {
			var exercise string
			var count int
			var createdAt time.Time
			err = rows.Scan(&exercise, &count, &createdAt)
			if err != nil {
				logError(nil, err, "unable to scan results for user's reps")
				return nil
			}
			for _, rd := range repDatas {
				// find which repData slot we need to populate. Probably more effecient way to do this. Probably a fancy mysql query could have done all this for me.
				if rd.Date != fmt.Sprintf("%d-%d", int(createdAt.Month()), createdAt.Day()) {
					continue
				}
				rd.ExerciseCounts[exercise] += count
			}
		}
		if rows.Err() != nil {
			logError(nil, rows.Err(), "error after parsing data for user reps")
		}
		officeReps[officeName] = repDatas
	}
	return officeReps
}

func initRepData() []RepData {
	var rd []RepData
	for cur := StartDate; cur.Before(EndDate.Add(time.Hour * 24)); cur = cur.Add(time.Hour * 24) {
		rd = append(
			rd, RepData{
				Date:           fmt.Sprintf("%d-%d", int(cur.Month()), cur.Day()),
				ExerciseCounts: make(map[string]int),
			})
	}
	return rd
}

func getUserReps(email string) []RepData {
	q := "SELECT reps.exercise, reps.count, reps.created_at FROM reps JOIN user on reps.user_id=user.id WHERE email=? AND reps.created_at > ? AND reps.created_at < ?"
	rows, err := DB.Query(q, email, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"))
	if err != nil {
		logError(nil, err, "unable to query for user's reps")
		return nil
	}
	repDatas := initRepData()
	for rows.Next() {
		var exercise string
		var count int
		var createdAt time.Time
		err = rows.Scan(&exercise, &count, &createdAt)
		if err != nil {
			logError(nil, err, "unable to scan results for user's reps")
			return nil
		}
		for _, rd := range repDatas {
			// find which repData slot we need to populate. Probably more effecient way to do this. Probably a fancy mysql query could have done all this for me.
			if rd.Date != fmt.Sprintf("%d-%d", int(createdAt.Month()), createdAt.Day()) {
				continue
			}
			rd.ExerciseCounts[exercise] += count
		}
	}
	if rows.Err() != nil {
		logError(nil, rows.Err(), "error after parsing data for user reps")
	}
	return repDatas
}

// mwPanic wraps the outer router so all panics are caught
func mwPanic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logEvent(r, "panic", fmt.Sprintf("%v %s", rec, debug.Stack()))
			}
		}()
		h.ServeHTTP(w, r)
	})
}

// logDataGet returns the log context
func logDataGet(r *http.Request) map[string]interface{} {
	if r == nil {
		return make(map[string]interface{})
	}
	ctx := r.Context()
	data := ctx.Value("log")
	switch v := data.(type) {
	case map[string]interface{}:
		return v
	}
	return make(map[string]interface{})
}

// logDataAdd adds a single value to the log context
func logDataAdd(r *http.Request, key string, value interface{}) {
	var data map[string]interface{}

	ctx := r.Context()
	d := ctx.Value("log")
	switch v := d.(type) {
	case map[string]interface{}:
		data = v
	default:
		data = make(map[string]interface{})
	}

	data[key] = value

	r = r.WithContext(context.WithValue(ctx, "log", data))
}

// logDataReplace replaces the current log context with the provided log data
func logDataReplace(r *http.Request, data map[string]interface{}) {
	ctx := r.Context()
	r = r.WithContext(context.WithValue(ctx, "log", data))
}

// ranOnce is a hack used when initializing the logger and mwLog
var ranOnce bool

// mwLog sets up the logger and caputues logging before the request returns
func mwLog(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logData := logDataGet(r)
		logData["request_time"] = start.Unix()
		logData["request_id"] = fmt.Sprintf("%08x", rand.Int63n(1e9))
		logData["event"] = "request"
		logData["remote_addr"] = r.RemoteAddr
		logData["method"] = r.Method
		logData["url"] = r.URL.String()
		logData["content_length"] = r.ContentLength

		// init the logger's response writer used to caputure the status code
		// pull from a pool, set the writer, initialize / reset the response code to a sensible default, reset that this response writer has been used
		// for the logging middleware (based on noodle's logger middleware)
		// could put the ranOnce in the init, but I want to make copy-pasta easier if I use mwLog again (before turning it into a real package)
		if !ranOnce {
			ranOnce = true
			writers.New = func() interface{} {
				return &logWriter{}
			}
		}
		lw := writers.Get().(*logWriter)
		lw.ResponseWriter = w
		lw.code = http.StatusOK
		lw.headerWritten = false
		defer writers.Put(lw)

		h.ServeHTTP(lw, r)

		logData["code"] = lw.Code()
		logData["tts_ns"] = time.Since(start).Nanoseconds() / 1e6 // time to serve in nano seconds

		log.Println(logAsString(logData))
	})
}

// logAsString returns the string version of the logs (ie, json marshal)
func logAsString(l map[string]interface{}) string {
	b, err := json.Marshal(l)
	if err != nil {
		logError(nil, err, "unable to marshal map[string]interface{}")
	}
	return string(b)
}

// logEvent allows us to track novel happenings
func logEvent(r *http.Request, event string, msg string) {
	logData := logDataGet(r)
	logData["event"] = event
	logData["message"] = msg

	log.Println(logAsString(logData))
}

// logError is similar to logEvent but has an error field
func logError(r *http.Request, err error, msg string) {
	logData := logDataGet(r)
	logData["event"] = "error"
	logData["message"] = msg
	if err == nil {
		err = fmt.Errorf("internal error condition")
	}
	logData["error"] = err.Error()

	log.Println(logAsString(logData))
}

// everything below is for the logger mw (from noodle)
// the purpose is to allow us to capture the response code that will be issued to the client

// logWriter mimics http.ResponseWriter functionality while storing
// HTTP status code for later logging
type logWriter struct {
	code          int
	headerWritten bool
	http.ResponseWriter
}

func (l *logWriter) WriteHeader(code int) {
	l.headerWritten = false
	if !l.headerWritten {
		l.ResponseWriter.WriteHeader(code)
		l.code = code
		l.headerWritten = true
	}
}

func (l *logWriter) Write(buf []byte) (int, error) {
	l.headerWritten = true
	return l.ResponseWriter.Write(buf)
}

func (l *logWriter) Code() int {
	return l.code
}

// provide other typical ResponseWriter methods
func (l *logWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return l.ResponseWriter.(http.Hijacker).Hijack()
}

func (l *logWriter) CloseNotify() <-chan bool {
	return l.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (l *logWriter) Flush() {
	l.ResponseWriter.(http.Flusher).Flush()
}

var writers sync.Pool
