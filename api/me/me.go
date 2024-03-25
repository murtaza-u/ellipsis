package me

import (
	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/internal/conf"
	"github.com/murtaza-u/ellipsis/internal/sqlc"

	"github.com/labstack/echo/v4"
)

type API struct {
	db  *sqlc.Queries
	key conf.Key
}

func New(db *sqlc.Queries, key conf.Key) API {
	return API{
		db:  db,
		key: key,
	}
}

func (a API) Register(app *echo.Echo) {
	auth := middleware.NewAuthMiddleware(a.db)

	grp := app.Group("/me", auth.Required)
	grp.GET("", a.ProfilePage, auth.AuthInfo)
	grp.POST("/change-password", a.ChangePassword)
	grp.GET("/session", a.SessionPage, auth.AuthInfo)

	grp.GET("/session/delete/:id", a.DeleteSessionPage)
	grp.POST("/session/delete", a.DeleteSession)
}
