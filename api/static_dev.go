//go:build dev
// +build dev

package api

import (
	"path/filepath"

	"github.com/labstack/echo/v4"
)

func Static(app *echo.Echo) error {
	app.Static("/static", filepath.Join("api", "static"))
	return nil
}
