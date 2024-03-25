package provider

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/api/util"
	"github.com/murtaza-u/ellipsis/internal/sqlc"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"

	"github.com/a-h/templ"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mileusna/useragent"
	"golang.org/x/oauth2"
)

type ProviderGoogle struct {
	oauth2.Config
	*oidc.Provider
	db *sqlc.Queries
}

type googleUser struct {
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func NewGoogleProvider(db *sqlc.Queries, c Credentials) (Provider, error) {
	p, err := oidc.NewProvider(
		context.Background(),
		"https://accounts.google.com",
	)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     p.Endpoint(),
		RedirectURL:  c.BaseURL + "/google/callback",
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return ProviderGoogle{
		Provider: p,
		Config:   conf,
		db:       db,
	}, nil
}

func (p ProviderGoogle) Login(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login - Google | Ellipsis",
				view.Error(
					"failed to get session",
					http.StatusExpectationFailed,
				),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	state, err := util.GenerateRandom(25)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login - Google | Ellipsis",
				view.Error(
					"failed to generate state",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	returnTo := c.QueryParam("return_to")
	if returnTo == "" {
		returnTo = "/me"
	}

	sess.Values["state"] = state
	sess.Values["return_to"] = returnTo
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login - Google | Ellipsis",
				view.Error(
					"failed to save to session",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	return c.Redirect(http.StatusSeeOther, p.AuthCodeURL(state))
}

func (p ProviderGoogle) Callback(c echo.Context) error {
	q := new(CallbackParams)
	if err := c.Bind(q); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"failed to parse query params",
					http.StatusExpectationFailed,
				),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	if q.Err != "" {
		msg := q.Err
		if q.ErrDesc != "" {
			msg += " - " + q.ErrDesc
		}
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(msg, http.StatusExpectationFailed),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	sess, err := session.Get("session", c)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"failed to get session",
					http.StatusExpectationFailed,
				),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	state, ok := sess.Values["state"].(string)
	if !ok {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"missing state in session",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	// prevent timing attacks on state
	if subtle.ConstantTimeCompare([]byte(state), []byte(q.State)) == 0 {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"invalid state",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	tkn, err := p.Exchange(c.Request().Context(), q.Code)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"failed to exchange code for token",
					http.StatusExpectationFailed,
				),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	rawIDTkn, ok := tkn.Extra("id_token").(string)
	if !ok {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"missing id_token field in oauth2 token",
					http.StatusExpectationFailed,
				),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	v := p.Verifier(&oidc.Config{ClientID: p.ClientID})
	_, err = v.Verify(c.Request().Context(), rawIDTkn)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"failed to verify id token",
					http.StatusExpectationFailed,
				),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	info, err := p.UserInfo(c.Request().Context(), newTokenSource(tkn))
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"failed to fetch user's info",
					http.StatusExpectationFailed,
				),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	user := new(googleUser)
	if err := info.Claims(user); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"failed to unmarshal user info claims",
					http.StatusExpectationFailed,
				),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	var userID int64

	exists := true
	dbUser, err := p.db.GetUserByEmail(c.Request().Context(), user.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Callback - Google | Ellipsis",
					view.Error(
						"database operation failed",
						http.StatusInternalServerError,
					),
				),
				Status: http.StatusInternalServerError,
			})
		}
		exists = false
	}

	if exists {
		userID = dbUser.ID
	}

	if !exists {
		var avatar sql.NullString
		if user.Picture != "" {
			avatar.String = user.Picture
			avatar.Valid = true
		}
		res, err := p.db.CreateUser(c.Request().Context(), sqlc.CreateUserParams{
			Email: user.Email,
			// AvatarUrl: avatar,
		})
		if err != nil {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Callback - Google | Ellipsis",
					view.Error(
						"database operation failed",
						http.StatusInternalServerError,
					),
				),
				Status: http.StatusInternalServerError,
			})
		}
		userID, err = res.LastInsertId()
		if err != nil {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Callback - Google | Ellipsis",
					view.Error(
						"database operation failed",
						http.StatusInternalServerError,
					),
				),
				Status: http.StatusInternalServerError,
			})
		}
	}

	sessID, err := util.GenerateRandom(25)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"failed to generate session ID",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	var ua useragent.UserAgent
	uaRaw := c.Request().Header.Get("User-Agent")
	if uaRaw != "" {
		ua = useragent.Parse(uaRaw)
	}
	var browser sql.NullString
	if b := util.BrowserFromUA(ua); b != "" {
		browser.String = b
		browser.Valid = true
	}
	var os sql.NullString
	if ua.OS != "" {
		os.String = ua.OS
		os.Valid = true
	}

	expiresAt := time.Now().Add(time.Hour * 4)

	_, err = p.db.CreateSession(c.Request().Context(), sqlc.CreateSessionParams{
		ID:        sessID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Browser:   browser,
		Os:        os,
	})
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	c.SetCookie(&http.Cookie{
		Name:     "auth_session",
		Value:    sessID,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		Path:     "/",
	})

	returnTo, ok := sess.Values["return_to"].(string)
	if !ok {
		returnTo = "/me"
	}

	isBoosted := c.Request().Header.Get("HX-Boosted") != ""
	if !isBoosted {
		return c.Redirect(http.StatusFound, returnTo)
	}

	r := c.Response()
	r.Header().Set("HX-Redirect", returnTo)

	// render empty template
	h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
	return h.Component.Render(c.Request().Context(), r)
}
