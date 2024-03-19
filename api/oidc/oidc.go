package oidc

import (
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
	dream *dream.Store
}

func New(db *sqlc.Queries, dream *dream.Store) API {
	return API{
		db:    db,
		dream: dream,
	}
}

func (a API) Register(app *echo.Echo) {
	app.GET("/.well-known/openid-configuration", a.configuration)
	app.GET("/authorize", a.authorize)
}
