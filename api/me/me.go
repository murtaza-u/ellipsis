package me

import (
	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/internal/sqlc"

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
	auth := middleware.NewAuthMiddleware(a.db)

	grp := app.Group("/me", auth.Required)
	grp.GET("", a.ProfilePage, auth.AuthInfo)
	grp.POST("/change-password", a.ChangePassword)
	grp.GET("/session", a.SessionPage, auth.AuthInfo)
	grp.DELETE("/session/:id", a.DeleteSession)
}
