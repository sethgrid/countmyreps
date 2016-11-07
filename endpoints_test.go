package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/sethgrid/countmyreps/integration"
)

// TODO: abstract this out for a standard test set up and teardown
func TestJsonEndpoint(t *testing.T) {
	tmpDB := fmt.Sprintf("countmyreps_test_%d_%d", time.Now().Unix(), rand.Intn(100))
	db := integration.SetupDB("root@tcp(127.0.0.1:3306)/?parseTime=true", tmpDB, true)
	integration.Seed(db, "2016-11-01", "2016-11-30", "2016-11-15")
	defer integration.TearDownDB(db, tmpDB)

	s := NewServer(db, 0, FakeEmailer{})
	go func() {
		err := s.Serve()
		if err != nil {
			t.Fatalf("unable to serve: %v", err)
		}
	}()
	defer s.Close()
	time.Sleep(1 * time.Second)
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/json?email=oc_1@sendgrid.com", s.Port))
	if err != nil {
		t.Fatalf("unable to get URL: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("error reading response body: %v", err)
	}

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Errorf("got %d, want %d for status code", got, want)
	}

	vd := ViewData{}
	err = json.Unmarshal(body, &vd)
	if err != nil {
		t.Error(err)
	}

	if got, want := vd.UserEmail, "oc_1@sendgrid.com"; got != want {
		t.Errorf("got %s, want %s for email", got, want)
	}
}
