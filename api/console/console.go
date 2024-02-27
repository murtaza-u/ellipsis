package console

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
	// overview
	app.GET("/console", a.overviewPage)

	// app
	app.GET("/console/app", a.appsPage)
	app.GET("/console/app/:id", a.appPage)
	app.GET("/console/app/create", a.createAppPage)
	app.POST("/console/app/create", a.createApp)
	app.POST("/console/app/:id/update", a.updateApp)
	app.DELETE("/console/app/:id/delete", a.deleteApp)

	// user
	app.GET("/console/user", a.userPage)
}
