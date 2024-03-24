package oidc

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type LogoutParams struct {
	IDTkn       string `query:"id_token_hint"`
	RedirectURI string `query:"post_logout_redirect_uri"`
	ClientID    string `query:"client_id"`
	State       string `query:"state"`
}

func (a API) Logout(c echo.Context) error {
	q := new(LogoutParams)
	if err := c.Bind(q); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Logout | Ellipsis",
				view.Error(
					"failed to parse query parameters",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}
	if q.IDTkn == "" {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Logout | Ellipsis",
				view.Error(
					"missing id_token_hint in query",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	claims := new(IDTknClaims)
	_, err := jwt.ParseWithClaims(
		q.IDTkn,
		claims,
		func(t *jwt.Token) (interface{}, error) {
			return *a.key.pub, nil
		},
	)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Logout | Ellipsis",
				view.Error(
					"failed to parse id_token",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	sess, err := a.DB.GetSessionWithClient(c.Request().Context(), claims.SID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Logout | Ellipsis",
					view.Error(
						"session does not exists",
						http.StatusBadRequest,
					),
				),
				Status: http.StatusBadRequest,
			})
		}
	}

	if q.ClientID != "" && q.ClientID != sess.ClientID {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Logout | Ellipsis",
				view.Error(
					"unauthorized client id",
					http.StatusUnauthorized,
				),
			),
			Status: http.StatusUnauthorized,
		})
	}

	q.RedirectURI = strings.TrimSpace(q.RedirectURI)
	q.RedirectURI = strings.TrimSuffix(q.RedirectURI, "/")

	var redirectTo string
	for _, u := range strings.Split(sess.LogoutCallbackUrls, ",") {
		if q.RedirectURI == "" || u == q.RedirectURI {
			redirectTo = u
			break
		}
	}

	if redirectTo == "" {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Logout | Ellipsis",
				view.Error(
					"Unauthorized redirect URI",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	err = a.DB.DeleteSession(c.Request().Context(), claims.SID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Logout | Ellipsis",
				view.Error(
					"database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	if q.State != "" {
		redirectTo += fmt.Sprintf("?state=%s", url.QueryEscape(q.State))
	}

	return c.Redirect(http.StatusFound, redirectTo)
}
