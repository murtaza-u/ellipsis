package me

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/oidc"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"
	"github.com/murtaza-u/ellipsis/view/partial/me"

	"github.com/a-h/templ"
	"github.com/golang-jwt/jwt/v5"
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

func (a API) DeleteSessionPage(c echo.Context) error {
	id := c.Param("id")
	sess, err := a.db.GetSessionWithOptionalClient(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Revoke Session | Ellipsis",
					view.Error(
						"invalid session id",
						http.StatusBadRequest,
					),
				),
				Status: http.StatusBadRequest,
			})
		}
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Revoke Session | Ellipsis",
				view.Error(
					"database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}
	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Revoke Session | Ellipsis",
			me.DeleteSession(sess.ClientName, sess.ID),
		),
	})
}

type deleteSessionParams struct {
	ID    string `form:"id"`
	Force bool   `form:"force"`
}

func (a API) DeleteSession(c echo.Context) error {
	form := new(deleteSessionParams)
	if err := c.Bind(form); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Revoke Session | Ellipsis",
				view.Error(
					"failed to parse form",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	if form.ID == "" {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Revoke Session | Ellipsis",
				view.Error(
					"missing session id",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	sess, err := a.db.GetSessionWithOptionalClient(c.Request().Context(), form.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Revoke Session | Ellipsis",
					view.Error(
						"invalid session id",
						http.StatusBadRequest,
					),
				),
				Status: http.StatusBadRequest,
			})
		}
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Revoke Session | Ellipsis",
				view.Error(
					"database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	if form.Force || !sess.ClientID.Valid {
		err := a.db.DeleteSession(c.Request().Context(), form.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return render.Do(render.Params{
					Ctx: c,
					Component: layout.Base(
						"Revoke Session | Ellipsis",
						view.Error(
							"invalid session id",
							http.StatusBadRequest,
						),
					),
					Status: http.StatusBadRequest,
				})
			}
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Revoke Session | Ellipsis",
					view.Error(
						"database operation failed",
						http.StatusInternalServerError,
					),
				),
				Status: http.StatusInternalServerError,
			})
		}

		isBoosted := c.Request().Header.Get("HX-Boosted") != ""
		if !isBoosted {
			return c.Redirect(http.StatusFound, "/me/session")
		}

		r := c.Response()
		r.Header().Set("HX-Redirect", "/me/session")

		// render empty template
		h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
		return h.Component.Render(c.Request().Context(), r)
	}

	if !sess.BackchannelLogoutUrl.Valid {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Revoke Session | Ellipsis",
				me.DeleteSessionUnsupportedBackchannelLogout(
					sess.ClientName.String,
					sess.ID,
				),
			),
		})
	}

	tkn, err := a.createLogoutTkn(sess.ID, sess.ClientID.String)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Revoke Session | Ellipsis",
				view.Error(
					"failed to generate logout token",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	err = a.backchannelLogout(
		c.Request().Context(),
		sess.BackchannelLogoutUrl.String,
		tkn,
	)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Revoke Session | Ellipsis",
				me.DeleteSessionBackchannelLogoutFailure(
					sess.ClientName.String,
					sess.ID,
				),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	err = a.db.DeleteSession(c.Request().Context(), form.ID)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Revoke Session | Ellipsis",
				view.Error(
					"database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	isBoosted := c.Request().Header.Get("HX-Boosted") != ""
	if !isBoosted {
		return c.Redirect(http.StatusFound, "/me/session")
	}

	r := c.Response()
	r.Header().Set("HX-Redirect", "/me/session")

	// render empty template
	h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
	return h.Component.Render(c.Request().Context(), r)
}

func (a API) createLogoutTkn(sid, clientID string) (string, error) {
	tkn := jwt.NewWithClaims(jwt.SigningMethodEdDSA, oidc.LogoutTknClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.baseURL,
			Subject:   clientID,
			Audience:  jwt.ClaimStrings{clientID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 2)),
		},
		Events: map[string]struct{}{
			"http://schemas.openid.net/event/backchannel-logout": {},
		},
		SID: sid,
	})
	return tkn.SignedString(a.key.Priv)
}

func (a API) backchannelLogout(ctx context.Context, uri, tkn string) error {
	q := make(url.Values)
	q.Set("logout_token", tkn)

	ctx, cancel := context.WithTimeout(ctx, time.Second*7)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		uri,
		strings.NewReader(q.Encode()),
	)
	if err != nil {
		return fmt.Errorf("failed to create new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call backchannel logout URI")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("client responded with non-200 status code: %s", resp.Status)
	}

	return nil
}
