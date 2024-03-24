package api

import (
	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"

	"github.com/labstack/echo/v4"
)

func (Server) indexPage(c echo.Context) error {
	var avatarURL string
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		avatarURL = ctx.AvatarURL
	}
	return render.Do(render.Params{
		Ctx:       c,
		Component: layout.Base("Ellipsis", view.Index(avatarURL)),
	})
}
