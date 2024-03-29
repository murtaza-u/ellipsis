package db

import (
	"fmt"
	"os"
)

type MySQLConfig struct {
	User     string
	Pass     string
	Database string
}

func (c MySQLConfig) validate() error {
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

type SqliteConfig struct {
	Path string
}

func (c SqliteConfig) validate() error {
	if c.Path == "" {
		return fmt.Errorf("missing database path")
	}
	info, err := os.Stat(c.Path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to access db %q", c.Path)
	}
	if info != nil && info.IsDir() {
		return fmt.Errorf("db path %q is a directory", c.Path)
	}
	return nil
}
