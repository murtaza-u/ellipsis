package oidc

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"path/filepath"

	"github.com/murtaza-u/account/api/middleware"
	"github.com/murtaza-u/account/api/oidc/provider"
	"github.com/murtaza-u/account/api/util"
	"github.com/murtaza-u/account/internal/conf"
	"github.com/murtaza-u/account/internal/sqlc"

	"github.com/labstack/echo/v4"
	"github.com/murtaza-u/dream"
)

const (
	ScopeOIDC    = "openid"
	ScopeProfile = "profile"
)

type API struct {
	Config
	key key
}

type Config struct {
	DB        *sqlc.Queries
	Cache     *dream.Store
	KeyStore  string
	Providers conf.Providers
}

type key struct {
	priv *ed25519.PrivateKey
	pub  *ed25519.PublicKey
}

func New(c Config) (*API, error) {
	data, err := os.ReadFile(filepath.Join(c.KeyStore, "ed25519"))
	if err != nil {
		return nil, fmt.Errorf("failed to read ed25519 priv key: %w", err)
	}
	priv, err := util.PEMToEd25519PrivKey(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read ed25519 priv key: %w", err)
	}

	data, err = os.ReadFile(filepath.Join(c.KeyStore, "ed25519.pub"))
	if err != nil {
		return nil, fmt.Errorf("failed to read ed25519 pub key: %w", err)
	}
	pub, err := util.PEMToEd25519PubKey(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read ed25519 pub key: %w", err)
	}

	return &API{
		Config: c,
		key: key{
			priv: priv,
			pub:  pub,
		},
	}, nil
}

func (a API) Register(app *echo.Echo) error {
	app.GET("/.well-known/openid-configuration", a.configuration)

	auth := middleware.NewAuthMiddleware(a.DB)
	app.GET("/authorize", a.authorize, auth.Required)
	app.POST("/authorize", a.consent, auth.Required)

	app.POST("/oauth/token", a.Token)
	app.GET("/.well-known/jwks.json", a.JWKs)
	app.GET("/userinfo", a.UserInfo)

	if a.Providers.Google.Enable {
		google, err := provider.NewGoogleProvider(a.DB, provider.Credentials{
			ClientID:     a.Providers.Google.ClientID,
			ClientSecret: a.Providers.Google.ClientSecret,
		})
		if err != nil {
			return fmt.Errorf("failed to setup google identity provider")
		}
		app.GET("/google/login", google.Login)
		app.GET("/google/callback", google.Callback)
	}

	if a.Providers.Github.Enable {
		github := provider.NewGithubProvider(a.DB, provider.Credentials{
			ClientID:     a.Providers.Github.ClientID,
			ClientSecret: a.Providers.Github.ClientSecret,
		})
		app.GET("/github/login", github.Login)
		app.GET("/github/callback", github.Callback)
	}

	return nil
}
