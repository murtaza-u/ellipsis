//go:build !dev
// +build !dev

package api

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed static
var staticFs embed.FS

func getFileSystem() (http.FileSystem, error) {
	fsys, err := fs.Sub(staticFs, "static")
	if err != nil {
		return nil, err
	}
	return http.FS(fsys), nil
}

func Static(app *echo.Echo) error {
	fs, err := getFileSystem()
	if err != nil {
		return err
	}
	h := http.FileServer(fs)
	app.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", h)))
	return nil
}
