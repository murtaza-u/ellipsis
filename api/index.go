package api

import (
	"github.com/labstack/echo/v4"

	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"
)

func (Server) index(c echo.Context) error {
	return render(c, layout.Base("Account", view.Index()))
}
