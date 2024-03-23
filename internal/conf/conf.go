package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type C struct {
	Port                 uint16    `yaml:"port"`
	KeyStore             string    `yaml:"keyStore"`
	SessionEncryptionKey string    `yaml:"sessionEncryptionKey"`
	Mysql                Mysql     `yaml:"mysql"`
	Providers            Providers `yaml:"providers"`
}

type Mysql struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type Providers struct {
	Google Provider `yaml:"google"`
	Github Provider `yaml:"github"`
}

type Provider struct {
	Enable       bool   `yaml:"enable"`
	ClientID     string `yaml:"clientID"`
	ClientSecret string `yaml:"clientSecret"`
}

func New(path string) (*C, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %q: %w", path, err)
	}

	c := new(C)
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config %q: %w", path, err)
	}

	return c, nil
}

func (c *C) Validate() error {
	if c.Port == 0 {
		c.Port = 3000
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

	if c.SessionEncryptionKey == "" {
		return fmt.Errorf("missing session encryption key")
	}

	if c.Mysql.User == "" {
		return fmt.Errorf("missing database user")
	}
	if c.Mysql.Password == "" {
		return fmt.Errorf("missing database password")
	}
	if c.Mysql.Database == "" {
		return fmt.Errorf("missing database")
	}

	var providers = []Provider{
		c.Providers.Google,
		c.Providers.Github,
	}
	for _, p := range providers {
		if !p.Enable {
			continue
		}
		if p.ClientID == "" {
			return fmt.Errorf("missing client ID for one of the enabled IdP")
		}
		if p.ClientSecret == "" {
			return fmt.Errorf("missing client secret for one of the enabled IdP")
		}
	}

	return nil
}
