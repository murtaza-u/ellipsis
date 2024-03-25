package oidc

import (
	"fmt"

	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/oidc/provider"
	"github.com/murtaza-u/ellipsis/internal/conf"
	"github.com/murtaza-u/ellipsis/internal/sqlc"

	"github.com/labstack/echo/v4"
	"github.com/murtaza-u/dream"
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
	Cache     *dream.Store
	Key       conf.Key
	Providers conf.Providers
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
