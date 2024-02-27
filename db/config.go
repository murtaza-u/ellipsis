package db

import "fmt"

type Config struct {
	User     string
	Pass     string
	Database string
}

func (c Config) validate() error {
	if c.User == "" {
		return fmt.Errorf("missing user")
	}
	if c.Pass == "" {
		return fmt.Errorf("missing password")
	}
	if c.Database == "" {
		return fmt.Errorf("missing database")
	}
	return nil
}
