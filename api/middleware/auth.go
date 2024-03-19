package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"net/url"

	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/internal/sqlc"
	"github.com/murtaza-u/account/view"

	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	db *sqlc.Queries
}

func NewAuthMiddleware(db *sqlc.Queries) AuthMiddleware {
	return AuthMiddleware{
		db: db,
	}
}

type CtxWithUserID struct {
	echo.Context
	UserID int64
}

func (m AuthMiddleware) Required(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		returnTo := url.QueryEscape(c.Request().URL.RequestURI())
		redirectTo := "/login?return_to=" + returnTo
		cookie, err := c.Cookie("session")
		if err != nil {
			return c.Redirect(http.StatusTemporaryRedirect, redirectTo)
		}
		sess, err := m.db.GetSession(c.Request().Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return c.Redirect(http.StatusTemporaryRedirect, redirectTo)
			}
			return render.Do(render.Params{
				Ctx: c,
				Component: view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
				Status: http.StatusInternalServerError,
			})
		}
		return next(CtxWithUserID{
			Context: c,
			UserID:  sess.UserID,
		})
	}
}

func (m AuthMiddleware) AlreadyAuthenticated(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		if err != nil {
			return next(c)
		}
		_, err = m.db.GetSession(c.Request().Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return next(c)
			}
			return render.Do(render.Params{
				Ctx: c,
				Component: view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
				Status: http.StatusInternalServerError,
			})
		}
		return c.Redirect(http.StatusTemporaryRedirect, "/me")
	}
}
