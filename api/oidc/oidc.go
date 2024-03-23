package oidc

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"path/filepath"

	"github.com/murtaza-u/account/api/middleware"
	"github.com/murtaza-u/account/api/util"
	"github.com/murtaza-u/account/internal/sqlc"
	"github.com/murtaza-u/dream"

	"github.com/labstack/echo/v4"
)

const (
	ScopeOIDC    = "openid"
	ScopeProfile = "profile"
)

type API struct {
	db    *sqlc.Queries
	cache *dream.Store
	key   key
}

type key struct {
	priv *ed25519.PrivateKey
	pub  *ed25519.PublicKey
}

func New(db *sqlc.Queries, cache *dream.Store, keyStore string) (*API, error) {
	data, err := os.ReadFile(filepath.Join(keyStore, "ed25519"))
	if err != nil {
		return nil, fmt.Errorf("failed to read ed25519 priv key: %w", err)
	}
	priv, err := util.PEMToEd25519PrivKey(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read ed25519 priv key: %w", err)
	}

	data, err = os.ReadFile(filepath.Join(keyStore, "ed25519.pub"))
	if err != nil {
		return nil, fmt.Errorf("failed to read ed25519 pub key: %w", err)
	}
	pub, err := util.PEMToEd25519PubKey(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read ed25519 pub key: %w", err)
	}

	return &API{
		db:    db,
		cache: cache,
		key: key{
			priv: priv,
			pub:  pub,
		},
	}, nil
}

func (a API) Register(app *echo.Echo) {
	app.GET("/.well-known/openid-configuration", a.configuration)

	auth := middleware.NewAuthMiddleware(a.db)
	app.GET("/authorize", a.authorize, auth.Required)
	app.POST("/authorize", a.consent, auth.Required)

	app.POST("/oauth/token", a.Token)
	app.GET("/.well-known/jwks.json", a.JWKs)
	app.GET("/userinfo", a.UserInfo)
}
