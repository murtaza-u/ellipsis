package render

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type Params struct {
	Ctx       echo.Context
	Component templ.Component
	Status    int
}

func Do(p Params) error {
	stat := http.StatusOK
	if p.Status != 0 {
		stat = p.Status
	}
	c := templ.Handler(p.Component, templ.WithStatus(stat)).Component
	return c.Render(p.Ctx.Request().Context(), p.Ctx.Response())
}
