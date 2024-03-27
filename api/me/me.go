package me

import (
	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/fs"
	"github.com/murtaza-u/ellipsis/internal/conf"
	"github.com/murtaza-u/ellipsis/internal/sqlc"

	"github.com/labstack/echo/v4"
)

type API struct {
	db      *sqlc.Queries
	key     conf.Key
	baseURL string
	fs      fs.Storage
}

func New(db *sqlc.Queries, key conf.Key, baseURL string, fs fs.Storage) API {
	return API{
		db:      db,
		key:     key,
		baseURL: baseURL,
		fs:      fs,
	}
}

func (a API) Register(app *echo.Echo) {
	auth := middleware.NewAuthMiddleware(a.db)

	grp := app.Group("", auth.Required, auth.AuthInfo)
	grp.GET("/", a.Profile)
	grp.POST("/", a.ChangeAvatar)
	grp.GET("/change-password", a.ChangePasswordPage)
	grp.POST("/change-password", a.ChangePassword)
	grp.GET("/session", a.SessionPage, auth.AuthInfo)
	grp.GET("/session/delete/:id", a.DeleteSessionPage)
	grp.POST("/session/delete", a.DeleteSession)
}
