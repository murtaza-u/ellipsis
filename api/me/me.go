package me

import (
	"github.com/murtaza-u/account/api/middleware"
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
	auth := middleware.NewAuthMiddleware(a.db)

	grp := app.Group("/me", auth.Required)
	grp.GET("", a.ProfilePage)
	grp.POST("/change-password", a.ChangePassword)
	grp.GET("/session", a.SessionPage)
	grp.DELETE("/session/:id", a.DeleteSession)
}
