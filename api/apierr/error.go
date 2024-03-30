package apierr

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/view"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

type Error struct {
	code int
	err  error
	comp templ.Component
}

func (e Error) Error() string {
	return fmt.Sprintf("%d - %s", e.code, e.err.Error())
}

func New(code int, err error, comp templ.Component) error {
	return &Error{
		code: code,
		err:  err,
		comp: comp,
	}
}

func Handler(err error, c echo.Context) {
	var (
		e  *Error
		ok bool
	)
	if e, ok = err.(*Error); !ok {
		return
	}

	code := e.code
	if code == 0 {
		code = http.StatusInternalServerError
	}

	err = e.err
	if err == nil {
		err = errors.New(http.StatusText(code))
	}

	comp := e.comp
	if comp == nil {
		comp = view.Error(err.Error(), code)
	}

	slog.Error(err.Error())

	render.Do(render.Params{
		Ctx:       c,
		Component: comp,
		Status:    code,
	})
}
