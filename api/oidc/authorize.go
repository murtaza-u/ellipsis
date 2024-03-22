package oidc

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/mileusna/useragent"
	"github.com/murtaza-u/account/api/middleware"
	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/api/util"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"

	"github.com/labstack/echo/v4"
)

type authorizeParams struct {
	ClientID           string `query:"client_id"`
	ResponseType       string `query:"response_type"`
	Scope              string `query:"scope"`
	State              string `query:"state"`
	RedirectURI        string `query:"redirect_uri"`
	IDTknSignedRespAlg string `query:"id_token_signed_response_alg"`
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

type authorizeMetadata struct {
	ClientID string   `json:"client_id"`
	UserID   int64    `json:"user_id"`
	Scopes   []string `json:"scopes"`
	Browser  string   `json:"browser"`
	OS       string   `json:"os"`
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

	redirectStat := http.StatusTemporaryRedirect

	if p.ResponseType != "code" {
		err := newAuthorizeErr("bad_request", "response type not supported")
		return c.Redirect(redirectStat, err.AttachTo(redirectTo))
	}

	if p.IDTknSignedRespAlg != "" && p.IDTknSignedRespAlg != "EdDSA" {
		err := newAuthorizeErr("bad_request",
			"unsupported id token signing algorithm")
		return c.Redirect(redirectStat, err.AttachTo(redirectTo))
	}

	scopes := strings.Split(p.Scope, " ")
	if len(scopes) == 0 {
		err := newAuthorizeErr("bad_request", "missing scope")
		return c.Redirect(redirectStat, err.AttachTo(redirectTo))
	}
	var hasOpenIDScope, hasInvalidScope bool
	for _, s := range scopes {
		if s == ScopeOIDC {
			hasOpenIDScope = true
		}
		if s != ScopeOIDC && s != ScopeProfile {
			hasInvalidScope = true
			break
		}
	}
	if hasInvalidScope {
		err := newAuthorizeErr("bad_request", "unsupported scope")
		return c.Redirect(redirectStat, err.AttachTo(redirectTo))
	}
	if !hasOpenIDScope {
		err := newAuthorizeErr("bad_request", "missing openid scope")
		return c.Redirect(redirectStat, err.AttachTo(redirectTo))
	}

	code, err := util.GenerateRandom(13)
	if err != nil {
		err := newAuthorizeErr("internal_server_error",
			"failed to generate authorization code")
		return c.Redirect(redirectStat, err.AttachTo(redirectTo))
	}

	var userID int64
	if ctx, ok := c.(middleware.CtxWithUserID); ok {
		userID = ctx.UserID
	}
	if userID == 0 {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Authorization - Account",
				view.Error(
					"Unauthorized user",
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
	var browser string
	if b := util.BrowserFromUA(ua); b != "" {
		browser = b
	}

	a.cache.Put(code, authorizeMetadata{
		ClientID: client.ID,
		UserID:   userID,
		Scopes:   scopes,
		OS:       ua.OS,
		Browser:  browser,
	})

	return c.Redirect(http.StatusFound, fmt.Sprintf(
		"%s?code=%s&state=%s",
		redirectTo,
		url.QueryEscape(code),
		url.QueryEscape(p.State),
	))
}
