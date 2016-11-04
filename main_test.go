package main

import (
	"strings"
	"testing"
)

func TestExtractEmailAddr(t *testing.T) {
	tests := []struct {
		email     string
		extracted string
	}{
		{"example@example.com", "example@example.com"},
		{"<example@example.com>", "example@example.com"},
		{"Example Email <example@example.com>", "example@example.com"},
	}
	for _, test := range tests {
		if got, want := extractEmailAddr(test.email), test.extracted; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestOfficeComparisonUpdateLeading(t *testing.T) {
	stats := fakeStats()
	msg := officeComparisonUpdate("RWC", stats)

	wanted := []string{
		"Your office is leading with 5 reps per day and 33% participating, with those Gridders doing 15 reps per day!",
	}

	notWanted := []string{
		"With a little effort, you can catch up",
	}

	t.Logf("Message body:\n%s", msg)

	for _, want := range wanted {
		if !strings.Contains(msg, want) {
			t.Errorf("Not found: %q", want)
		}
	}

	for _, notWant := range notWanted {
		if strings.Contains(msg, notWant) {
			t.Errorf("Found, but not wanted: %q", notWant)
		}
	}
}

func TestOfficeComparisonUpdateNotLeading(t *testing.T) {
	stats := fakeStats()
	msg := officeComparisonUpdate("OC", stats)

	wanted := []string{
		"Your office has 2 reps per day and 10% participating, with those Gridders doing 10 reps per day.",
		"With a little effort, you can catch up to the RWC office who are doing 5 reps per day, and have 33% particpating",
	}

	notWanted := []string{
		"Your office is leading",
	}

	t.Logf("Message body:\n%s", msg)

	for _, want := range wanted {
		if !strings.Contains(msg, want) {
			t.Errorf("Not found: %q", want)
		}
	}

	for _, notWant := range notWanted {
		if strings.Contains(msg, notWant) {
			t.Errorf("Found, but not wanted: %q", notWant)
		}
	}
}

func fakeStats() map[string]Stats {
	stats := make(map[string]Stats)

	RWCStats := Stats{}
	RWCStats.OfficeSize = 100
	RWCStats.PercentParticipating = 33
	RWCStats.RepsPerPerson = 14
	RWCStats.RepsPerPersonParticipating = 45
	RWCStats.RepsPerPersonParticipatingPerDay = 15
	RWCStats.RepsPerPersonPerDay = 5
	RWCStats.TotalReps = 495
	stats["RWC"] = RWCStats

	OCStats := Stats{}
	OCStats.OfficeSize = 200
	OCStats.PercentParticipating = 10
	OCStats.RepsPerPerson = 3
	OCStats.RepsPerPersonParticipating = 30
	OCStats.RepsPerPersonParticipatingPerDay = 10
	OCStats.RepsPerPersonPerDay = 2
	OCStats.TotalReps = 600
	stats["OC"] = OCStats

	DenverStats := Stats{}
	DenverStats.OfficeSize = 300
	DenverStats.PercentParticipating = 5
	DenverStats.RepsPerPerson = 33
	DenverStats.RepsPerPersonParticipating = 75
	DenverStats.RepsPerPersonParticipatingPerDay = 25
	DenverStats.RepsPerPersonPerDay = 4
	DenverStats.TotalReps = 1000
	stats["Denver"] = DenverStats

	return stats
}
