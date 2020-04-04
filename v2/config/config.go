package config

import "fmt"

type Config struct {
	UseHTTPS           bool   `envconfig:"use_https" default:"true"`
	Addr               string `envconfig:"addr" default:"countmyreps.com"` // use localhost:5000 in local dev
	Port               int    `envconfig:"port" default:"5000"`
	DBPath             string `envconfig:"db_path" default:"./cmr.db"`
	RemoveDBOnShutdown bool   `envconfig:"remove_db_on_shutdown" default:"false"`
	GoogleCredsPath    string `envconfig:"google_creds_path" default:"../../creds.json"`

	// computed
	FullAddr string
}

// Sanitize will clean up fixable config errors and error out on validation problems
func (c *Config) Sanitize() error {
	if c.DBPath == "" {
		return fmt.Errorf("db_path cannot be empty")
	}

	scheme := "https"
	if !c.UseHTTPS {
		scheme = "http"
	}
	c.FullAddr = fmt.Sprintf("%s://%s", scheme, c.Addr)
	return nil
}
