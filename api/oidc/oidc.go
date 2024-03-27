package oidc

import (
	"fmt"

	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/oidc/provider"
	"github.com/murtaza-u/ellipsis/fs"
	"github.com/murtaza-u/ellipsis/internal/conf"
	"github.com/murtaza-u/ellipsis/internal/sqlc"

	"github.com/labstack/echo/v4"
)

const (
	ScopeOIDC    = "openid"
	ScopeProfile = "profile"
)

type API struct {
	Config
}

type Config struct {
	DB        *sqlc.Queries
	Key       conf.Key
	Providers conf.Providers
	BaseURL   string
	FS        fs.Storage
}

func New(c Config) (*API, error) {
	return &API{Config: c}, nil
}

func (a API) Register(app *echo.Echo) error {
	app.GET("/.well-known/openid-configuration", a.configuration)

	auth := middleware.NewAuthMiddleware(a.DB)
	app.GET("/authorize", a.authorize, auth.Required, auth.AuthInfo)
	app.POST("/authorize", a.consent, auth.Required, auth.AuthInfo)

	app.POST("/oauth/token", a.Token)
	app.GET("/.well-known/jwks.json", a.JWKs)
	app.GET("/userinfo", a.UserInfo)
	app.GET("/oidc/logout", a.Logout)

	if a.Providers.Google.Enable {
		google, err := provider.NewGoogleProvider(
			a.DB, a.FS,
			provider.Credentials{
				ClientID:     a.Providers.Google.ClientID,
				ClientSecret: a.Providers.Google.ClientSecret,
				BaseURL:      a.BaseURL,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to setup google identity provider")
		}
		app.GET("/google/login", google.Login)
		app.GET("/google/callback", google.Callback)
	}

	if a.Providers.Github.Enable {
		github := provider.NewGithubProvider(
			a.DB, a.FS,
			provider.Credentials{
				ClientID:     a.Providers.Github.ClientID,
				ClientSecret: a.Providers.Github.ClientSecret,
				BaseURL:      a.BaseURL,
			},
		)
		app.GET("/github/login", github.Login)
		app.GET("/github/callback", github.Callback)
	}

	return nil
}
