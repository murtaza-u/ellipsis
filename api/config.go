package api

import "fmt"

type Config struct {
	Port             uint16
	DatabaseUser     string
	DatabasePassword string
	Database         string
}

func (c *Config) validate() error {
	if c.Port == 0 {
		c.Port = 3000
	}
	if c.DatabaseUser == "" {
		return fmt.Errorf("missing database user")
	}
	if c.DatabasePassword == "" {
		return fmt.Errorf("missing database password")
	}
	if c.Database == "" {
		return fmt.Errorf("missing database")
	}
	return nil
}
