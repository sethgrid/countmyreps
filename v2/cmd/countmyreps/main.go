package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
	countmyreps "github.com/sethgrid/countmyreps/v2"
	"github.com/sethgrid/countmyreps/v2/config"
)

func main() {

	c := &config.Config{}

	if err := envconfig.Process("countmyreps", c); err != nil {
		log.Fatalf("config error: unable to start countmyreps - %s", err)
	}

	s, err := countmyreps.NewServer(c)
	if err != nil {
		log.Fatalf("unable to create server: %s", err)
	}

	if err := s.Serve(); err != nil {
		log.Fatalf("unable to continue serving countmyreps: %s", err)
	}
}
