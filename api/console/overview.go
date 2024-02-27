package console

import (
	"net/http"

	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"
	"github.com/murtaza-u/account/view/partial/console"

	"github.com/labstack/echo/v4"
)

func (a API) overviewPage(c echo.Context) error {
	count, err := a.db.GetUserAndClientCount(c.Request().Context())
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Console | Account",
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
			"Console | Account",
			view.Console(
				"/console",
				console.Overview(
					int(count.ClientCount),
					int(count.UserCount),
				),
			),
		),
	})
}
