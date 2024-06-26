package console

import (
	"fmt"
	"net/http"

	"github.com/murtaza-u/ellipsis/api/apierr"
	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"
	"github.com/murtaza-u/ellipsis/view/partial/console"

	"github.com/labstack/echo/v4"
)

func (a API) userPage(c echo.Context) error {
	users, err := a.db.GetUsers(c.Request().Context())
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to read user from db: %w", err),
			layout.Base(
				"Console - Users | Ellipsis",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
		)
	}

	var avatarURL string
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		avatarURL = ctx.AvatarURL
	}

	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Console - Users | Ellipsis",
			view.Console("/console/user", avatarURL, console.Users(users)),
		),
	})
}
