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
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/facebookgo/flagenv"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

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

// Offices is all the valid Offices
var Offices []string

// AppName is the app name
var AppName = "countmyreps"

// Version is the semver
var Version = "2.0.0"

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
// TODO: it would be interesting to figure out how to have this dynamic (display too) and based off the email they send to
const (
	PullUps  = "Pull Ups"
	SitUps   = "Sit Ups"
	PushUps  = "Push Ups"
	Squats   = "Squats"
	OldEmail = "pullups-pushups-airsquats-situps@countmyreps.com"
	NewEmail = "pullups-pushups-squats-situps@countmyreps.com"
)

// TODO: secondary grouping aside from office. Like department or team or multiple.

// Debug turns on more verbose logging
var Debug bool

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
	flag.BoolVar(&Debug, "debug", false, "set flag for verbose logging")

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

	err = populateOfficesVar()
	if err != nil {
		log.Fatal(err)
	}

	// set up routes and serve
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/view", viewHandler)
	r.HandleFunc("/parseapi/index.php", parseHandler)                                  // backwards compatibility
	r.PathPrefix("/").Handler(http.StripPrefix("", http.FileServer(http.Dir("web/")))) // mux specific workaround for fileserver; todo: use separate mux to avoid filtering these endpoints from logs?

	http.Handle("/", mwPanic(mwLog(r)))

	log.Printf("starting on :%d", port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Println("Unexpected error serving: ", err.Error())
	}
}

func totalReps(d []RepData) int {
	sum := 0
	for _, rd := range d {
		for _, count := range rd.ExerciseCounts {
			sum += count
		}
	}
	return sum
}

func populateOfficesVar() error {
	q := "SELECT name FROM office"
	rows, err := DB.Query(q)
	if err != nil {
		return err
	}
	for rows.Next() {
		var office string
		err = rows.Scan(&office)
		if err != nil {
			return err
		}
		Offices = append(Offices, office)
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	return nil
}

// errorHandler is a helper method to log and display errors. When invoked from a parent handler, the parent should then return
func errorHandler(w http.ResponseWriter, r *http.Request, code int, message string, err error) {
	logError(r, err, message)
	w.WriteHeader(code)
	w.Write([]byte(fmt.Sprintf("%v - %s", http.StatusText(code), message)))
}

// SendErrorEmail sets up the error message and then calls sendEmail
func SendErrorEmail(rcpt string, originalAddressTo string, subject string, msg string) error {
	officeList := strings.Join(Offices, ", ")
	msgFmt := `
	<h3>Uh oh!</h3>
	<p>
	There was an error with your CountMyReps Submission.<br /><br />
    Make sure that you addressed your email to %s<br />
    Make sure that your subject line was FOUR comma separated numbers, like: 5, 10, 15, 20<br />
    If you were trying to set your office location, make sure you choose one from:<br />
	%s<br />
	(This should be sent in its own email).
    </p>
	<p>
    Details from received message:<br />
    Addessed to: %s<br />
    Subject: %s<br />
    Time: %s<br />
	Error: %s<br />
	</p>`
	return sendEmail(rcpt, "Error with your submission", fmt.Sprintf(msgFmt, NewEmail, officeList, originalAddressTo, subject, time.Now().String(), msg))
}

func officeComparisonUpdate(userOffice string, officeStats map[string]Stats) string {
	var leadOffice string
	var currentLeadCount int
	for office, stats := range officeStats {
		if stats.RepsPerPersonParticipatingPerDay >= currentLeadCount {
			leadOffice = office
		}
	}
	var msg string
	officePercent := fmt.Sprintf("%d%%", officeStats[userOffice].PercentParticipating)
	officePerDay := officeStats[userOffice].RepsPerPersonParticipatingPerDay
	if userOffice == leadOffice {
		msg = fmt.Sprintf("Your office is leading with %s%% participating, with those Gridders doing %d reps per day!", officePercent, officePerDay)
	} else {
		msg = fmt.Sprintf("Your office has %s%% participating, with those Gridders doing %d reps per day. With a little effort, you can catch up to the %s office who have %d%% particpating, doing %d reps per day.",
			officePercent, officePerDay,
			leadOffice,
			officeStats[leadOffice].PercentParticipating, officeStats[leadOffice].RepsPerPersonParticipatingPerDay)
	}
	return msg
}

// SendSuccessEmail sets up the success message and calls sendEmail
func SendSuccessEmail(to string) error {
	office := getUserOffice(to)
	officeStats := getOfficeStats()
	var officeMsg string
	var forTheTeam string
	if office == "" || office == "Unknown" {
		officeMsg = fmt.Sprintf("You've not linked your reps to an office. Send an email to %s with your office in the subject line. Valid office choices are: <br />%s", NewEmail, strings.Join(Offices, ", "))
		forTheTeam = ""
	} else {
		officeMsg = officeComparisonUpdate(office, officeStats)
		forTheTeam = fmt.Sprintf(" for the %s team", office)
	}
	total := totalReps(getUserReps(to))
	days := int(time.Since(StartDate).Hours() / float64(24))
	if days == 0 {
		days = 1 // avoid divide by zero
	}
	avg := total / days

	var data []string
	for officeName, stats := range officeStats {
		data = append(data, fmt.Sprintf("%s: %d", officeName, stats.TotalReps))
	}

	officeTotals := "The office totals are: " + strings.Join(data, ", ")

	msg := fmt.Sprintf(`<h3>Keep it up!</h3>
	<p>
	You've logged a total of %d%s, an average of %d per day.
	</p>
	<p>
	%s
	</p>
	<p>
	%s
	</p>`, total, forTheTeam, avg, officeMsg, officeTotals)

	return sendEmail(to, "Success!", fmt.Sprintf(msg))
}

func sendEmail(to string, subject string, msg string) error {
	from := mail.NewEmail("CountMyReps", "automailer@countmyreps.com")
	// at this point, all recipients _should_ be firstname.lastname@sendgrid.com or firstname@sendgrid.com
	toName := strings.Split(to, ".")[0]
	if strings.Contains(toName, "@") {
		toName = strings.Split(toName, "@")[0]
	}
	toAddr := mail.NewEmail(toName, to)

	msg = `<img src="http://countmyreps.com/images/mustache-thin.jpg" style="margin:auto; width:300px; display:block"/>` + msg

	content := mail.NewContent("text/html", msg)
	m := mail.NewV3MailInit(from, subject, toAddr, content)

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}
	if !(response.StatusCode == http.StatusOK || response.StatusCode == http.StatusAccepted) {
		return fmt.Errorf("unexpected status code from SendGrid: %d - %q", response.StatusCode, response.Body)
	}
	return nil
}

// ErrSubjectFmt ...
var ErrSubjectFmt = "CountMyReps was unable to parse your subject. Please provide FOUR comma separated numbers like: `5, 10, 15, 20` where the numbers represent pull ups, push ups, squats, and situps respectively. You provided \"%s\""

// ErrToAddrFmt ...
var ErrToAddrFmt = "CountMyReps only accepts emails to " + NewEmail + ", you sent to \"%s\""

// ErrFromFmt ...
var ErrFromFmt = "CountMyReps only accepts mail from the sendgrid domain. You used \"%s\""

// ErrUnexpectedFmt ...
var ErrUnexpectedFmt = "CountMyReps experienced an unexpected error, please try again later. Error: %s"

func parseHandler(w http.ResponseWriter, r *http.Request) {
	// NOTE: SendGrid's Inbound Parse API requires a 200 level response always, even on error, otherwise it will retry

	// errMsg is parsed later to determine if we should send a success or error email
	var errMsg string
	var err error

	to := r.PostFormValue("to")
	from := r.PostFormValue("from")
	subject := r.PostFormValue("subject")
	logDebug(r, fmt.Sprintf("from: %s; subject: %s, to: %s", from, subject, to))

	defer func() {
		var mailType string
		if errMsg != "" {
			mailType = "error - " + errMsg
			err = SendErrorEmail(from, to, subject, errMsg)
		} else {
			mailType = "success"
			err = SendSuccessEmail(from)
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

	userID, err := getOrCreateUserID(from)
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
			// TODO: move out of loop and use VALUES (), (), (), ()
			_, err = DB.Exec("INSERT INTO reps (exercise, count, user_id) VALUES (?, ?, ?)", exercise, count, userID)
			if err != nil {
				logError(r, err, "unable to insert rep")
				errMsg = fmt.Sprintf(ErrUnexpectedFmt, "unable to insert into the database")
				return
			}
		}
	} else if inListCaseInsenitive(subject, Offices) {
		office := formattedOffice(subject)
		_, err = DB.Exec("UPDATE user SET office=(SELECT id FROM office where name=?) WHERE id=? LIMIT 1", office, userID)
		if err != nil {
			logError(r, err, "unable to update user's office")
			errMsg = fmt.Sprintf(ErrUnexpectedFmt, "unable to update office relationship in the database")
			return
		}
	} else {
		logEvent(r, "bad_parse", fmt.Sprintf("bad subject: %s", subject))
		errMsg = fmt.Sprintf(ErrSubjectFmt, subject)
		return
	}
}

func formattedOffice(s string) string {
	for _, office := range Offices {
		if strings.ToLower(office) == strings.TrimSpace(strings.ToLower(s)) {
			return office
		}
	}
	logError(nil, fmt.Errorf("unable to determine office"), fmt.Sprintf("attempting to set office to %q", s))
	return "Unknown"
}

func inListCaseInsenitive(s string, list []string) bool {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	for _, elem := range list {
		if s == strings.ToLower(elem) {
			return true
		}
	}
	return false
}

func getOrCreateUserID(email string) (int, error) {
	var id int
	getQ := "SELECT id FROM user WHERE email=? LIMIT 1"
	row := DB.QueryRow(getQ, email)
	err := row.Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return 0, errors.Wrap(err, queryPrinter(getQ, email))
	} else if err == sql.ErrNoRows {
		q := "INSERT INTO user (email, office) VALUES (?, (SELECT id from office where name=\"\"))"
		res, err := DB.Exec(q, email)
		if err != nil {
			return 0, errors.Wrap(err, queryPrinter(q, email))
		}
		i, err := res.LastInsertId()
		if err != nil {
			return 0, err
		}
		id = int(i)
		logEvent(nil, "new_user", email)
	}
	return id, nil
}

// extractEmailAddr gets the email address from the email string
// John <Smith@example.com>
// <Smith@example.com>
// smith@example.com
// ^^ all gitve smith@example.com
func extractEmailAddr(email string) string {
	if !strings.Contains(email, "<") {
		return email
	}
	var extracted []rune
	var capture bool
	for _, r := range email {
		if string(r) == "<" {
			capture = true
			continue
		}
		if string(r) == ">" {
			capture = false
			continue
		}
		if capture {
			extracted = append(extracted, r)
		}
	}
	return string(extracted)
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

// getTodaysReps will only grab the latest N submissions
func getTodaysReps(email string) []RepData {
	var rd []RepData
	limit := 11
	q := fmt.Sprintf("SELECT reps.exercise, reps.count, reps.created_at FROM reps JOIN user on reps.user_id=user.id WHERE user.email=? AND created_at >= ? ORDER BY created_at DESC LIMIT %d", limit)
	rows, err := DB.Query(q, email, fmt.Sprintf("%d-%d-%d", time.Now().Year(), int(time.Now().Month()), time.Now().Day()))
	if err != nil {
		logError(nil, errors.Wrap(err, queryPrinter(q, email, fmt.Sprintf("%d-%d-%d", time.Now().Year(), int(time.Now().Month()), time.Now().Day()))), "unable to get today's reps")
		return rd
	}
	for rows.Next() {
		var exercise string
		var count int
		var createdAt time.Time
		err := rows.Scan(&exercise, &count, &createdAt)
		if err != nil {
			logError(nil, errors.Wrap(err, queryPrinter(q, email, fmt.Sprintf("%d-%d-%d", time.Now().Year(), int(time.Now().Month()), time.Now().Day()))), "unable to scan today's reps")
		}
		rd = append(rd, RepData{
			Date:           createdAt.Format(time.Kitchen),
			ExerciseCounts: map[string]int{exercise: count},
		})
	}
	// reverse the data for presentation needs
	for i, j := 0, len(rd)-1; i < j; i, j = i+1, j-1 {
		rd[i], rd[j] = rd[j], rd[i]
	}
	return rd
}

func getUserOffice(email string) string {
	// leverage the empty value; there is a "" value in the office table
	var officeName string
	q := "SELECT office.name FROM user JOIN office ON user.office=office.id WHERE user.email=?"
	row := DB.QueryRow(q, email)
	err := row.Scan(&officeName)
	if err != nil && err != sql.ErrNoRows {
		logError(nil, errors.Wrap(err, queryPrinter(q, email)), "unable to query for office name")
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
			logError(nil, errors.Wrap(err, queryPrinter(qHeadCount, officeName)), "unable to scan for office head count")
		}

		qParticip := "SELECT count(distinct id) from (SELECT user.id FROM reps JOIN user on reps.user_id=user.id JOIN office ON user.office=office.id WHERE office.name=? and reps.created_at > ? AND reps.created_at < ?) participating;"
		row = DB.QueryRow(qParticip, officeName, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"))
		err = row.Scan(&participating)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			logError(nil, errors.Wrap(err, queryPrinter(qParticip, officeName, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"))), "unable to scan for office participation")
			return officeStats
		}

		qTotals := "select sum(reps.count) from reps left join user on reps.user_id=user.id join office on office.id=user.office where reps.created_at > ? and reps.created_at < ? and office.name=?;"
		row = DB.QueryRow(qTotals, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"), officeName)
		err = row.Scan(&totalReps)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			logError(nil, errors.Wrap(err, queryPrinter(qTotals, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"), officeName)), "unable to scan for office totals")
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
			logError(nil, errors.Wrap(err, queryPrinter(q, officeName, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"))), "unable to query for user's reps")
			return nil
		}
		repDatas := initRepData()
		for rows.Next() {
			var exercise string
			var count int
			var createdAt time.Time
			err = rows.Scan(&exercise, &count, &createdAt)
			if err != nil {
				logError(nil, errors.Wrap(err, queryPrinter(q, officeName, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"))), "unable to scan results for user's reps")
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
		logError(nil, errors.Wrap(err, queryPrinter(q, email, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"))), "unable to query for user's reps")
		return nil
	}
	repDatas := initRepData()
	for rows.Next() {
		var exercise string
		var count int
		var createdAt time.Time
		err = rows.Scan(&exercise, &count, &createdAt)
		if err != nil {
			logError(nil, errors.Wrap(err, queryPrinter(q, email, StartDate.Format("2006-01-02"), EndDate.Format("2006-01-02"))), "unable to scan results for user's reps")
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
	l["app"] = AppName
	l["version"] = Version
	b, err := json.Marshal(l)
	if err != nil {
		log.Printf("unable to marshap map[string]interface{}. Wtf. %v \n %#v", err, l)
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

// logDebug captures debug messages and only passes them through if Debug is true
func logDebug(r *http.Request, msg string) {
	if Debug {
		logEvent(r, "debug", msg)
	}
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
