package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/sethgrid/countmyreps/integration"
)

func setup() *Server {
	tmpDB := fmt.Sprintf("countmyreps_test_%d_%d", time.Now().Unix(), rand.Intn(100))
	db := integration.SetupDB("root@tcp(127.0.0.1:3306)/?parseTime=true", tmpDB, true)
	integration.Seed(db, "2016-11-01", "2016-11-30", "2016-11-15")

	s := NewServer(db, 0, FakeEmailer{})
	s.dbname = tmpDB
	go func() {
		err := s.Serve()
		if err != nil {
			log.Fatalf("unable to serve: %v", err)
		}
	}()

	for s.Port == 0 {
		time.Sleep(1 * time.Millisecond)
	}

	err := populateOfficesVar(db)
	if err != nil {
		log.Fatal(err)
	}

	return s
}

func teardown(s *Server) {
	s.Close()
	integration.TearDownDB(s.DB, s.dbname)
	s.DB.Close()
}

type distilledResponse struct {
	code int
	body []byte
}

func getResponse(port int, path string) (*distilledResponse, error) {
	fullURL := fmt.Sprintf("http://127.0.0.1:%d%s", port, path)
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to GET %s", fullURL)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read resp body")
	}

	return &distilledResponse{
		code: resp.StatusCode,
		body: body,
	}, nil
}

func TestJsonEndpoint(t *testing.T) {
	srv := setup()
	defer teardown(srv)

	// oc_1@sendgrid.com is a known user from integration.Seed()
	resp, err := getResponse(srv.Port, "/json?email=oc_1@sendgrid.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Resp Body:\n%s", resp.body)

	if got, want := resp.code, http.StatusOK; got != want {
		t.Errorf("got %d, want %d for status code", got, want)
	}

	vd := ViewData{}
	err = json.Unmarshal(resp.body, &vd)
	if err != nil {
		t.Error(err)
	}

	if got, want := vd.UserEmail, "oc_1@sendgrid.com"; got != want {
		t.Errorf("got %s, want %s for email", got, want)
	}
}

func TestViewEndpoint(t *testing.T) {
	srv := setup()
	defer teardown(srv)

	// oc_1@sendgrid.com is a known user from integration.Seed()
	resp, err := getResponse(srv.Port, "/view?email=oc_1@sendgrid.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Resp Body:\n%s", resp.body)

	if got, want := resp.code, http.StatusOK; got != want {
		t.Errorf("got %d, want %d for status code", got, want)
	}

	for _, want := range []string{
		// user email should be on the page
		`oc_1@sendgrid.com`,
		// user teams should be on the page
		`eng`,
		`crossfit`,
		`mp`,
		// offices should be on the page
		`OC`,
		`Denver`,
	} {
		if !bytes.Contains(resp.body, []byte(want)) {
			t.Errorf("did not find in body:\n%s", want)
		}
	}

	for _, notWant := range []string{
		`unable to execute`,
		`Internal Server Error`,
		`view.html`,
	} {
		if bytes.Contains(resp.body, []byte(notWant)) {
			t.Errorf("found (and don't want) in body:\n%s", notWant)
		}
	}

}

func TestIndex(t *testing.T) {
	srv := setup()
	defer teardown(srv)

	resp, err := getResponse(srv.Port, "/")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Resp Body:\n%s", resp.body)

	if got, want := resp.code, http.StatusOK; got != want {
		t.Errorf("got %d, want %d for status code", got, want)
	}

	for _, want := range []string{
		`form action="view" method="get"`,
		`Send your email to pullups-pushups-squats-situps@countmyreps.com`,
	} {
		if !bytes.Contains(resp.body, []byte(want)) {
			t.Errorf("did not find in body:\n%s", want)
		}
	}
}
