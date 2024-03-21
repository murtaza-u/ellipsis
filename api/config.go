package api

import (
	"fmt"
	"os"
)

type Config struct {
	Port             uint16
	DatabaseUser     string
	DatabasePassword string
	Database         string
	KeyStore         string
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
	if c.KeyStore == "" {
		return fmt.Errorf("missing key store")
	}
	info, err := os.Stat(c.KeyStore)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("key store %q does not exist", c.KeyStore)
		}
		return fmt.Errorf("failed to access key store")
	}
	if !info.IsDir() {
		return fmt.Errorf("key store %q is not a directory", c.KeyStore)
	}
	return nil
}
