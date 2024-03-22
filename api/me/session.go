package me

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/murtaza-u/account/api/middleware"
	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"
	"github.com/murtaza-u/account/view/partial/me"
)

func (a API) SessionPage(c echo.Context) error {
	var (
		userID int64
		sessID string
	)
	if ctx, ok := c.(middleware.CtxWithIDs); ok {
		userID = ctx.UserID
		sessID = ctx.SessionID
	}
	if userID == 0 || sessID == "" {
		r := c.Response()
		r.Header().Set("HX-Redirect", "/logout")

		// render empty template
		h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
		return h.Component.Render(c.Request().Context(), r)
	}

	sessions, err := a.db.GetSessionForUserID(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r := c.Response()
			r.Header().Set("HX-Redirect", "/logout")

			// render empty template
			h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
			return h.Component.Render(c.Request().Context(), r)
		}
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"My Account - Sessions | Account",
				view.Me(
					"/me/session",
					view.Error(
						"database operation failed",
						http.StatusInternalServerError,
					),
				),
			),
			Status: http.StatusInternalServerError,
		})
	}
	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"My Account - Sessions | Account",
			view.Me(
				"/me/session",
				me.Sessions(sessions, sessID),
			),
		),
	})
}

func (a API) DeleteSession(c echo.Context) error {
	id := c.Param("id")
	err := a.db.DeleteSession(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: view.Error(
					"invalid session id",
					http.StatusBadRequest,
				),
				Status: http.StatusBadRequest,
			})
		}
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"database operation failed",
				http.StatusInternalServerError,
			),
			Status: http.StatusInternalServerError,
		})
	}
	return render.Do(render.Params{
		Ctx:       c,
		Component: view.Empty(),
	})
}
