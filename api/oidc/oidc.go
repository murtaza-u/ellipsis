package oidc

import (
	"github.com/murtaza-u/account/api/middleware"
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
}

func New(db *sqlc.Queries, cache *dream.Store) API {
	return API{
		db:    db,
		cache: cache,
	}
}

func (a API) Register(app *echo.Echo) {
	app.GET("/.well-known/openid-configuration", a.configuration)

	auth := middleware.NewAuthMiddleware(a.db)
	app.GET("/authorize", a.authorize, auth.Required)
}
