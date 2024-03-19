package oidc

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/api/util"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"

	"github.com/labstack/echo/v4"
)

type authorizeParams struct {
	ClientID     string `query:"client_id"`
	ResponseType string `query:"response_type"`
	Scope        string `query:"scope"`
	State        string `query:"state"`
	RedirectURI  string `query:"redirect_uri"`
}

type authorizeErr struct {
	name string
	desc string
}

func newAuthorizeErr(name, desc string) authorizeErr {
	return authorizeErr{
		name: name,
		desc: desc,
	}
}

func (a authorizeErr) Error() string {
	return fmt.Sprintf("%s - %s", a.name, a.desc)
}

func (a authorizeErr) Name() string {
	return url.QueryEscape(a.name)
}

func (a authorizeErr) Desc() string {
	return url.QueryEscape(a.desc)
}

func (a authorizeErr) AttachTo(baseURL string) string {
	return fmt.Sprintf(
		"%s?error=%s&error_description=%s",
		baseURL, a.Name(), a.Desc())
}

type clientIDWithScopes struct {
	ID     string   `json:"id"`
	Scopes []string `json:"scopes"`
}

func (a API) authorize(c echo.Context) error {
	p := new(authorizeParams)
	if err := c.Bind(p); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Authorization - Account",
				view.Error(
					"Failed to parse query parameters",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	client, err := a.db.GetClient(c.Request().Context(), p.ClientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Authorization - Account",
					view.Error(
						"Invalid client id",
						http.StatusBadRequest,
					),
				),
				Status: http.StatusBadRequest,
			})
		}
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Authorization - Account",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	p.RedirectURI = strings.TrimSpace(p.RedirectURI)
	p.RedirectURI = strings.TrimSuffix(p.RedirectURI, "/")

	var redirectTo string
	for _, u := range strings.Split(client.CallbackUrls, ",") {
		if p.RedirectURI == "" || u == p.RedirectURI {
			redirectTo = u
			break
		}
	}

	if redirectTo == "" {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Authorization - Account",
				view.Error(
					"Unauthorized redirect URI",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	if p.ResponseType != "code" {
		err := newAuthorizeErr("bad_request", "response type not supported")
		return c.Redirect(http.StatusTemporaryRedirect, err.AttachTo(redirectTo))
	}

	scopes := strings.Split(p.Scope, " ")
	if len(scopes) == 0 {
		err := newAuthorizeErr("bad_request", "missing scope")
		return c.Redirect(http.StatusTemporaryRedirect, err.AttachTo(redirectTo))
	}

	var validScope bool
	for _, s := range scopes {
		if s == ScopeOIDC || s == ScopeProfile {
			validScope = true
			break
		}
	}
	if !validScope {
		err := newAuthorizeErr("bad_request", "unsupported scope")
		return c.Redirect(http.StatusTemporaryRedirect, err.AttachTo(redirectTo))
	}

	code, err := util.GenerateRandom(13)
	if err != nil {
		err := newAuthorizeErr("internal_server_error",
			"failed to generate authorization code")
		return c.Redirect(http.StatusTemporaryRedirect, err.AttachTo(redirectTo))
	}

	v, err := json.Marshal(clientIDWithScopes{
		ID:     client.ID,
		Scopes: scopes,
	})
	if err != nil {
		err := newAuthorizeErr("internal_server_error",
			"database operation failed")
		return c.Redirect(http.StatusTemporaryRedirect, err.AttachTo(redirectTo))
	}
	a.dream.Put(code, v)

	return c.Redirect(http.StatusFound, fmt.Sprintf(
		"%s?code=%s&state=%s",
		redirectTo,
		url.QueryEscape(code),
		url.QueryEscape(p.State),
	))
}
