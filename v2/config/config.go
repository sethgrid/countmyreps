package config

import (
	"fmt"
	"log"
	"strings"
)

type Config struct {
	UseHTTPS           bool   `envconfig:"use_https" default:"true"`
	Addr               string `envconfig:"addr" default:"countmyreps.com"` // use localhost:5000 in local dev
	Port               int    `envconfig:"port" default:"5000"`
	DBPath             string `envconfig:"db_path" default:"./cmr.db"`
	RemoveDBOnShutdown bool   `envconfig:"remove_db_on_shutdown" default:"false"`
	GoogleCredsPath    string `envconfig:"google_creds_path" default:"../../creds.json"`

	// When set to true, the server will not contact Google OAuth2. Instead, the handler will take the passed in `code` and store that as the user's email address
	DevMode bool `envconfig:"dev_mode" default:"false"`
	// computed
	FullAddr string
}

// Sanitize will clean up fixable config errors and error out on validation problems
func (c *Config) Sanitize() error {
	if c.DBPath == "" {
		return fmt.Errorf("db_path cannot be empty")
	}

	if c.DevMode && !strings.Contains(c.Addr, "localhost") {
		log.Println("dev mode disabled when addr is not localhost")
		c.DevMode = false
	}

	scheme := "https"
	if !c.UseHTTPS {
		scheme = "http"
	}
	c.FullAddr = fmt.Sprintf("%s://%s", scheme, c.Addr)
	return nil
}
