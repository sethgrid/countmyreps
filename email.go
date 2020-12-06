package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// TODO: change to interface so we can unit test

// ErrSubjectFmt ...
var ErrSubjectFmt = "CountMyReps was unable to parse your subject. Please provide FOUR comma separated numbers like: `5, 10, 15, 20` where the numbers represent pull ups, push ups, squats, and situps respectively. You provided \"%s\""

// ErrToAddrFmt ...
var ErrToAddrFmt = "CountMyReps only accepts emails to " + NewEmail + ", you sent to \"%s\""

// ErrFromFmt ...
var ErrFromFmt = "CountMyReps only accepts mail from the sendgrid domain. You used \"%s\""

// ErrUnexpectedFmt ...
var ErrUnexpectedFmt = "CountMyReps experienced an unexpected error, please try again later. Error: %s"

// Emailer interface allows us to send emails
type Emailer interface {
	SendEmail(to string, subject string, msg string, replyMessageID string) error
}

// FakeEmailer is useful for testing
type FakeEmailer struct {
	Err error
}

// SendEmail is a NoOp for the FakeEmailer, returning what ever error we need
func (f FakeEmailer) SendEmail(to string, subject string, msg string) error {
	return f.Err
}

// SendGridEmailer matches the Emailer interface allowing us to send email through SendGrid
type SendGridEmailer struct{}

// SendEmail sends an email through SendGrid
func (SendGridEmailer) SendEmail(to string, subject string, msg string, replyMessageID string) error {
	from := mail.NewEmail("CountMyReps", "automailer@countmyreps.partkyle.io")
	// at this point, all recipients _should_ be firstname.lastname@sendgrid.com or firstname@sendgrid.com
	toName := strings.Split(to, ".")[0]
	if strings.Contains(toName, "@") {
		toName = strings.Split(toName, "@")[0]
	}
	toAddr := mail.NewEmail(toName, to)

	msg = `<img src="https://countmyreps.partkyle.io/images/mustache-thin.jpg" style="margin:auto; width:300px; display:block"/>` + msg

	content := mail.NewContent("text/html", msg)
	m := mail.NewV3MailInit(from, subject, toAddr, content)
	if replyMessageID != "" {
		m.SetHeader("In-Reply-To", replyMessageID)
	}

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

// SendErrorEmail sets up the error message and then calls sendEmail
func (s *Server) SendErrorEmail(rcpt string, originalAddressTo string, subject string, msg string, replyMessageID string) error {
	officeList := strings.Join(Offices, ", ")
	msgFmt := `
	<h3>Uh oh!</h3>
	<p>
	There was an error with your CountMyReps Submission.<br /><br />
    Make sure that you addressed your email to %s<br />
    Make sure that your subject line was FOUR comma separated numbers, like: 5, 10, 15, 20<br />
    If you were trying to set your office location, make sure you choose one from:<br />
	%s<br />
	(This should be sent in its own email). The same for if you are removing or adding a team. Use 'Team Add: team-name' or 'Team Remove: team-name'.
    </p>
	<p>
    Details from received message:<br />
    Addessed to: %s<br />
    Subject: %s<br />
    Time: %s<br />
	Error: %s<br />
	</p>`
	return EmailSender.SendEmail(rcpt, subject, fmt.Sprintf(msgFmt, NewEmail, officeList, originalAddressTo, subject, time.Now().String(), msg), replyMessageID)
}

// SendSuccessEmail sets up the success message and calls sendEmail
func (s *Server) SendSuccessEmail(to string, subject string, replyMessageID string) error {
	office := getUserOffice(s.DB, to)
	officeStats := getOfficeStats(s.DB)
	var officeMsg string
	var forTheTeam string
	if office == "" || office == "Unknown" {
		officeMsg = fmt.Sprintf("You've not linked your reps to an office. Send an email to %s with your office in the subject line. Valid office choices are: <br />%s", NewEmail, strings.Join(Offices, ", "))
		forTheTeam = ""
	} else {
		officeMsg = officeComparisonUpdate(office, officeStats)
		forTheTeam = fmt.Sprintf(" for the %s team", office)
	}
	total := totalReps(getUserReps(s.DB, to))
	days := int(time.Since(StartDate).Hours() / float64(24))
	if days == 0 {
		days = 1 // avoid divide by zero
	}
	avg := total / days

	var data []string
	for officeName, stats := range officeStats {
		data = append(data, fmt.Sprintf("%s: %d", officeName, stats.TotalReps))
	}

	var teamsMsg string
	teams := getUserTeams(s.DB, to)
	if len(teams) == 0 {
		teamsMsg = "You are not with any teams yet! Send an email with the subject 'Team Add: team-name' to get on a team. You can be on multiple teams!"
	} else {
		teamsMsg = "You are on the following teams. You can send an email with the subject 'Team Remove: team-name' to get off them, or `Team Add: team-name' to join others."
		var teamList string
		for _, team := range teams {
			teamList += fmt.Sprintf("<li>%s</li>", team)
		}
		teamsMsg = fmt.Sprintf("%s<ul>%s</ul>", teamsMsg, teamList)
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
	</p>
	<p>
	%s
	</p>`, total, forTheTeam, avg, officeMsg, teamsMsg, officeTotals)

	return EmailSender.SendEmail(to, subject, msg, replyMessageID)
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

func officeComparisonUpdate(userOffice string, officeStats map[string]Stats) string {
	var leadOffice string
	var currentLeadCount int
	for office, stats := range officeStats {
		if stats.RepsPerPersonPerDay >= currentLeadCount {
			leadOffice = office
			currentLeadCount = stats.RepsPerPersonPerDay
		}
	}
	var msg string
	if userOffice == leadOffice {
		msg = fmt.Sprintf("Your office is leading with %d reps per day and %d%% participating, with those Gridders doing %d reps per day!",
			officeStats[userOffice].RepsPerPersonPerDay,
			officeStats[userOffice].PercentParticipating,
			officeStats[userOffice].RepsPerPersonParticipatingPerDay,
		)
	} else {
		msg = fmt.Sprintf("Your office has %d reps per day and %d%% participating, with those Gridders doing %d reps per day. With a little effort, you can catch up to the %s office who are doing %d reps per day, and have %d%% particpating",
			officeStats[userOffice].RepsPerPersonPerDay,
			officeStats[userOffice].PercentParticipating,
			officeStats[userOffice].RepsPerPersonParticipatingPerDay,
			leadOffice,
			officeStats[leadOffice].RepsPerPersonPerDay,
			officeStats[leadOffice].PercentParticipating,
		)
	}
	return msg
}
