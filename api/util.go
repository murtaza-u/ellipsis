package api

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func render(ctx echo.Context, c templ.Component, status ...int) error {
	stat := http.StatusOK
	if len(status) != 0 {
		stat = status[0]
	}
	c = templ.Handler(c, templ.WithStatus(stat)).Component
	return c.Render(ctx.Request().Context(), ctx.Response())
}
