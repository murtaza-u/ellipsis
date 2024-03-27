package conf

import (
	"crypto/ed25519"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/murtaza-u/ellipsis/api/util"

	"gopkg.in/yaml.v3"
)

type C struct {
	BaseURL              string    `yaml:"baseURL"`
	Port                 uint16    `yaml:"port"`
	KeyStore             string    `yaml:"keyStore"`
	SessionEncryptionKey string    `yaml:"sessionEncryptionKey"`
	Mysql                Mysql     `yaml:"mysql"`
	Providers            Providers `yaml:"providers"`
	S3                   S3        `yaml:"s3"`

	Key Key
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

type S3 struct {
	Bucket string `yaml:"bucket"`
	Region string `yaml:"region"`
}

type Key struct {
	Priv ed25519.PrivateKey
	Pub  ed25519.PublicKey
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
	if c.BaseURL == "" {
		return fmt.Errorf("missing base url")
	}
	if _, err := url.ParseRequestURI(c.BaseURL); err != nil {
		return fmt.Errorf("invalid base url %q", c.BaseURL)
	}
	c.BaseURL = strings.TrimSuffix(c.BaseURL, "/")

	if c.Port == 0 {
		c.Port = 3000
	}

	if c.SessionEncryptionKey == "" {
		return fmt.Errorf("missing session encryption key")
	}

	if err := c.readKeysFromStore(); err != nil {
		return err
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

	if c.S3.Bucket == "" {
		return fmt.Errorf("missing s3 bucket")
	}
	if c.S3.Region == "" {
		return fmt.Errorf("missing s3 region")
	}

	return nil
}

func (c *C) readKeysFromStore() error {
	// check if `keystore` directory exists
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

	// read and parse private key
	data, err := os.ReadFile(filepath.Join(c.KeyStore, "ed25519"))
	if err != nil {
		return fmt.Errorf("failed to read ed25519 priv key: %w", err)
	}

	priv, err := util.PEMToEd25519PrivKey(data)
	if err != nil {
		return fmt.Errorf("failed to read ed25519 priv key: %w", err)
	}

	// read and parse public key
	data, err = os.ReadFile(filepath.Join(c.KeyStore, "ed25519.pub"))
	if err != nil {
		return fmt.Errorf("failed to read ed25519 pub key: %w", err)
	}
	pub, err := util.PEMToEd25519PubKey(data)
	if err != nil {
		return fmt.Errorf("failed to read ed25519 pub key: %w", err)
	}

	c.Key = Key{Priv: priv, Pub: pub}

	return nil
}
