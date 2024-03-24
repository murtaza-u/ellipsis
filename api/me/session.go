package me

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"
	"github.com/murtaza-u/ellipsis/view/partial/me"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

func (a API) SessionPage(c echo.Context) error {
	var (
		userID            int64
		sessID, avatarURL string
	)
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		userID = ctx.UserID
		sessID = ctx.SessionID
		avatarURL = ctx.AvatarURL
	}
	if userID == 0 || sessID == "" {
		r := c.Response()
		r.Header().Set("HX-Redirect", "/logout")

		// render empty template
		h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
		return h.Component.Render(c.Request().Context(), r)
	}

	sess, err := a.db.GetSessionWithClientForUserID(c.Request().Context(), userID)
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
				"My Account - Sessions | Ellipsis",
				view.Me(
					"/me/session",
					avatarURL,
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
			"My Account - Sessions | Ellipsis",
			view.Me(
				"/me/session",
				avatarURL,
				me.Sessions(sess, sessID),
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
