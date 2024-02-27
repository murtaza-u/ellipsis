package oidc

import (
	"github.com/murtaza-u/account/internal/sqlc"

	"github.com/labstack/echo/v4"
)

type API struct {
	db *sqlc.Queries
}

func New(db *sqlc.Queries) API {
	return API{
		db: db,
	}
}

func (a API) Register(app *echo.Echo) {
	app.GET("/.well-known/openid-configuration", a.configuration)
	app.GET("/authorize", a.authorize)
}
