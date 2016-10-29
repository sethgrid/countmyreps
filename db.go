package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

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
		if office != "" {
			Offices = append(Offices, office)
		}
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	return nil
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

func queryPrinter(q string, args ...interface{}) string {
	qFmt := strings.Replace(q, "?", `"%v"`, -1)
	return fmt.Sprintf(qFmt, args...)
}
