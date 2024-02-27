package console

import (
	"net/http"

	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"
	"github.com/murtaza-u/account/view/partial/console"

	"github.com/labstack/echo/v4"
)

func (a API) userPage(c echo.Context) error {
	users, err := a.db.GetUsers(c.Request().Context())
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Console - Users | Account",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}
	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Console - Users | Account",
			view.Console(
				"/console/user",
				console.Users(users),
			),
		),
	})
}
