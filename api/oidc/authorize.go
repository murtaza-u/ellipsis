package oidc

import (
	"net/http"

	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"

	"github.com/labstack/echo/v4"
)

type authorizeParams struct {
	ClientID     string `query:"client_id"`
	ResponseType string `query:"responseType"`
	Scope        string `query:"scope"`
	State        string `query:"state"`
	RedirectURI  string `query:"redirect_uri"`
}

func (a API) authorize(c echo.Context) error {
	p := new(authorizeParams)
	if err := c.Bind(p); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Authorization - Account",
				view.Error(
					"Failed to parse query parameters",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}
	return nil
}
