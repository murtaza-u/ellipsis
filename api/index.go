package api

import (
	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"

	"github.com/labstack/echo/v4"
)

func (Server) indexPage(c echo.Context) error {
	return render.Do(render.Params{
		Ctx:       c,
		Component: layout.Base("Account", view.Index()),
	})
}
