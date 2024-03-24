package oidc

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/api/util"
	"github.com/murtaza-u/ellipsis/internal/sqlc"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mileusna/useragent"
)

func (a API) consent(c echo.Context) error {
	isBoosted := c.Request().Header.Get("HX-Boosted") != ""

	form := new(consentParams)
	if err := c.Bind(form); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"failed to parse form",
				http.StatusBadRequest,
			),
			Status: http.StatusBadRequest,
		})
	}

	if form.ReturnTo == "" {
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"missing returning URL",
				http.StatusBadRequest,
			),
			Status: http.StatusBadRequest,
		})
	}
	if form.Callback == "" {
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"missing callback URL",
				http.StatusBadRequest,
			),
			Status: http.StatusBadRequest,
		})
	}

	if form.Consent != "granted" {
		err := newAuthorizeErr("denied", "user did not consent")
		callback := err.AttachTo(form.Callback)

		if !isBoosted {
			return c.Redirect(http.StatusTemporaryRedirect, callback)
		}

		r := c.Response()
		r.Header().Set("HX-Redirect", callback)

		// render empty template
		h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
		return h.Component.Render(c.Request().Context(), r)
	}

	client, err := a.DB.GetClient(c.Request().Context(), form.ClientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: view.Error(
					"Invalid client id",
					http.StatusBadRequest,
				),
				Status: http.StatusBadRequest,
			})
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

	var userID int64
	if ctx, ok := c.(middleware.CtxWithIDs); ok {
		userID = ctx.UserID
	}

	_, err = a.DB.CreateAuthzHistory(
		c.Request().Context(),
		sqlc.CreateAuthzHistoryParams{
			UserID:   userID,
			ClientID: client.ID,
		},
	)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"Database operation failed",
				http.StatusInternalServerError,
			),
			Status: http.StatusInternalServerError,
		})
	}

	if !isBoosted {
		return c.Redirect(http.StatusFound, form.ReturnTo)
	}

	r := c.Response()
	r.Header().Set("HX-Redirect", form.ReturnTo)

	// render empty template
	h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
	return h.Component.Render(c.Request().Context(), r)
}

type consentParams struct {
	Consent  string `form:"consent"`
	Callback string `form:"callback"`
	ReturnTo string `form:"return_to"`
	ClientID string `form:"client_id"`
}

func (a API) authorize(c echo.Context) error {
	p := new(authorizeParams)
	if err := c.Bind(p); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Authorization | Ellipsis",
				view.Error(
					"Failed to parse query parameters",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	client, err := a.DB.GetClient(c.Request().Context(), p.ClientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Authorization | Ellipsis",
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
				"Authorization | Ellipsis",
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
				"Authorization | Ellipsis",
				view.Error(
					"Unauthorized redirect URI",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	var userID int64
	if ctx, ok := c.(middleware.CtxWithIDs); ok {
		userID = ctx.UserID
	}
	if userID == 0 {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Authorization | Ellipsis",
				view.Error(
					"An internal error occured",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}
	u, err := a.DB.GetUser(c.Request().Context(), userID)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Authorization | Ellipsis",
				view.Error(
					"An internal error occured",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	_, err = a.DB.GetAuthzHistory(
		c.Request().Context(),
		sqlc.GetAuthzHistoryParams{
			UserID:   u.ID,
			ClientID: client.ID,
		},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Authorize | Ellipsis",
					view.Authorize(redirectTo, c.Request().RequestURI, u, client),
				),
			})
		}
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Authorization | Ellipsis",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
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
	var hasOpenIDScope, hasProfileScope, hasInvalidScope bool
	for _, s := range scopes {
		if s == ScopeOIDC {
			hasOpenIDScope = true
			continue
		}
		if s == ScopeProfile {
			hasProfileScope = true
			continue
		}
		hasInvalidScope = true
		break
	}
	if hasInvalidScope {
		err := newAuthorizeErr("bad_request", "unsupported scope")
		return c.Redirect(redirectStat, err.AttachTo(redirectTo))
	}
	if !hasOpenIDScope {
		err := newAuthorizeErr("bad_request", "missing openid scope")
		return c.Redirect(redirectStat, err.AttachTo(redirectTo))
	}
	if !hasProfileScope {
		err := newAuthorizeErr("bad_request", "missing profile scope")
		return c.Redirect(redirectStat, err.AttachTo(redirectTo))
	}

	code, err := util.GenerateRandom(13)
	if err != nil {
		err := newAuthorizeErr("internal_server_error",
			"failed to generate authorization code")
		return c.Redirect(redirectStat, err.AttachTo(redirectTo))
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

	a.Cache.Put(code, authorizeMetadata{
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
