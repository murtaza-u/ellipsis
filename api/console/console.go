package console

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

	grp := app.Group("/console", auth.Required)

	// overview
	grp.GET("", a.overviewPage)

	// app
	grp.GET("/app", a.appsPage)
	grp.GET("/app/:id", a.appPage)
	grp.PUT("/app/:id", a.updateApp)
	grp.DELETE("/app/:id", a.deleteApp)
	grp.GET("/app/create", a.createAppPage)
	grp.POST("/app/create", a.createApp)

	// user
	grp.GET("/user", a.userPage)
}
