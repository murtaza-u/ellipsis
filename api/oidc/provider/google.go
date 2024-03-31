package provider

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/murtaza-u/ellipsis/api/apierr"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/api/util"
	"github.com/murtaza-u/ellipsis/fs"
	"github.com/murtaza-u/ellipsis/internal/sqlc"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"

	"github.com/a-h/templ"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

type ProviderGoogle struct {
	oauth2.Config
	*oidc.Provider
	db *sqlc.Queries
	fs fs.Storage
}

type googleUser struct {
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func NewGoogleProvider(db *sqlc.Queries, fs fs.Storage, c Credentials) (Provider, error) {
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
		fs:       fs,
	}, nil
}

func (p ProviderGoogle) Login(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return apierr.New(
			http.StatusExpectationFailed,
			fmt.Errorf("failed to read session cookie store: %w", err),
			layout.Base(
				"Login - Google | Ellipsis",
				view.Error(
					"an internal error occured",
					http.StatusExpectationFailed,
				),
			),
		)
	}

	state, err := util.GenerateRandom(25)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to generate random string: %w", err),
			layout.Base(
				"Login - Google | Ellipsis",
				view.Error(
					"failed to generate state",
					http.StatusInternalServerError,
				),
			),
		)
	}

	returnTo := c.QueryParam("return_to")
	if returnTo == "" {
		returnTo = "/"
	}

	sess.Values["state"] = state
	sess.Values["return_to"] = returnTo
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to save to session cookie: %w", err),
			layout.Base(
				"Login - Google | Ellipsis",
				view.Error(
					"an internal error occured",
					http.StatusInternalServerError,
				),
			),
		)
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
		return apierr.New(
			http.StatusExpectationFailed,
			fmt.Errorf("failed to read session cookie store: %w", err),
			layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"an internal error occured",
					http.StatusExpectationFailed,
				),
			),
		)
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
		return apierr.New(
			http.StatusExpectationFailed,
			fmt.Errorf("[Google] failed to exchange code for token: %w", err),
			layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"failed to exchange code for token",
					http.StatusExpectationFailed,
				),
			),
		)
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

	userID, err := p.InsertUser(c.Request().Context(), user)
	if err != nil {
		return apierr.New(
			http.StatusExpectationFailed,
			fmt.Errorf("failed to insert user into db: %w", err),
			layout.Base(
				"Callback - GitHub | Ellipsis",
				view.Error(
					"database operation failed",
					http.StatusInternalServerError,
				),
			),
		)
	}

	sessID, err := util.GenerateRandom(25)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to generate random string: %w", err),
			layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"failed to generate session ID",
					http.StatusInternalServerError,
				),
			),
		)
	}

	fingerprint := util.ParseUA(c.Request().Header.Get("User-Agent"))
	expiresAt := time.Now().Add(time.Hour * 4)

	_, err = p.db.CreateSession(c.Request().Context(), sqlc.CreateSessionParams{
		ID:        sessID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Browser:   fingerprint.Browser,
		Os:        fingerprint.OS,
	})
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to insert session into db: %w", err),
			layout.Base(
				"Callback - Google | Ellipsis",
				view.Error(
					"database operation failed",
					http.StatusInternalServerError,
				),
			),
		)
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
		returnTo = "/"
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

func (p ProviderGoogle) InsertUser(ctx context.Context, u *googleUser) (string, error) {
	dbUser, err := p.db.GetUserByEmail(ctx, u.Email)

	// user exists
	if err == nil {
		return dbUser.ID, nil
	}

	// database error
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("failed to read user: %w", err)
	}

	// user does not exists
	userID, err := util.GenerateRandom(25)
	if err != nil {
		return "", fmt.Errorf("failed to generate user id: %w", err)
	}

	f, err := util.ReadURL(ctx, u.Picture)
	err = p.fs.Put(userID, f)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to file storage: %w", err)
	}

	url, err := p.fs.GetURL(userID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch file url from storage: %w", err)
	}

	_, err = p.db.CreateUser(ctx, sqlc.CreateUserParams{
		ID:        userID,
		Email:     u.Email,
		AvatarUrl: sql.NullString{String: url, Valid: true},
	})
	if err != nil {
		return "", fmt.Errorf("failed to insert new user: %w", err)
	}

	return userID, nil
}
