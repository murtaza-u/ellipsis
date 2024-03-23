package provider

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/api/util"
	"github.com/murtaza-u/account/internal/sqlc"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"

	"github.com/a-h/templ"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mileusna/useragent"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type ProviderGithub struct {
	*oauth2.Config
	db *sqlc.Queries
}

const githubUserInfoURL = "https://api.github.com/user"

type githubUser struct {
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func NewGithubProvider(db *sqlc.Queries, c Credentials) Provider {
	conf := &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  "http://localhost:3000/github/callback",
		Scopes:       []string{"read:user"},
	}
	return ProviderGithub{
		Config: conf,
		db:     db,
	}
}

func (p ProviderGithub) Login(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login - GitHub | Account",
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
				"Login - GitHub | Account",
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
				"Login - GitHub | Account",
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

func (p ProviderGithub) Callback(c echo.Context) error {
	q := new(CallbackParams)
	if err := c.Bind(q); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - GitHub | Account",
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
				"Callback - GitHub | Account",
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
				"Callback - GitHub | Account",
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
				"Callback - GitHub | Account",
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
				"Callback - GitHub | Account",
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
				"Callback - GitHub | Account",
				view.Error(
					"failed to exchange code for token",
					http.StatusExpectationFailed,
				),
			),
			Status: http.StatusExpectationFailed,
		})
	}

	req, err := http.NewRequest(http.MethodGet, githubUserInfoURL, nil)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - GitHub | Account",
				view.Error(
					fmt.Sprintf("failed to retreive user: %s", err.Error()),
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tkn.AccessToken))

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - GitHub | Account",
				view.Error(
					fmt.Sprintf("failed to retreive user: %s", err.Error()),
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	if resp.StatusCode != http.StatusOK {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - GitHub | Account",
				view.Error(
					fmt.Sprintf("failed to retreive user: status: %s", resp.Status),
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - GitHub | Account",
				view.Error(
					fmt.Sprintf("failed to retreive user: %s", err.Error()),
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	user := new(githubUser)
	err = json.Unmarshal(body, user)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Callback - GitHub | Account",
				view.Error(
					fmt.Sprintf("failed to retreive user: %s", err.Error()),
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
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
					"Callback - GitHub | Account",
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
		if user.AvatarURL != "" {
			avatar.String = user.AvatarURL
			avatar.Valid = true
		}
		res, err := p.db.CreateUser(c.Request().Context(), sqlc.CreateUserParams{
			Email:     user.Email,
			AvatarUrl: avatar,
		})
		if err != nil {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Callback - GitHub | Account",
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
					"Callback - GitHub | Account",
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
				"Callback - GitHub | Account",
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
				"Callback - GitHub | Account",
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
