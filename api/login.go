package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/api/util"
	"github.com/murtaza-u/account/internal/sqlc"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"

	"github.com/a-h/templ"
	"github.com/alexedwards/argon2id"
	"github.com/labstack/echo/v4"
	"github.com/mileusna/useragent"
)

func (Server) LoginPage(c echo.Context) error {
	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Login | Account",
			view.Login(view.LoginParams{
				ReturnTo: c.QueryParam("return_to"),
			}, map[string]error{}),
		),
	})
}

func (s Server) Login(c echo.Context) error {
	params := new(view.LoginParams)
	if err := c.Bind(params); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login | Account",
				view.Error("Failed to parse form", http.StatusBadRequest),
			),
			Status: http.StatusBadRequest,
		})
	}
	params.ReturnTo = c.QueryParam("return_to")

	errMap := make(map[string]error)

	if err := validateEmail(params.Email); err != nil {
		errMap["email"] = err
	}
	if params.Password == "" {
		errMap["password"] = errors.New("password can not be blank")
	}

	if len(errMap) != 0 {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login | Account",
				view.Login(*params, errMap),
			),
			Status: http.StatusBadRequest,
		})
	}

	u, err := s.queries.GetUserByEmail(c.Request().Context(), params.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errMap["email"] = errors.New("invalid E-Mail or password")
			errMap["password"] = errMap["email"]
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Login | Account",
					view.Login(*params, errMap),
				),
				Status: http.StatusBadRequest,
			})
		}
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login | Account",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}
	if !u.HashedPassword.Valid {
		errMap["email"] = errors.New("invalid E-Mail or password")
		errMap["password"] = errMap["email"]
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login | Account",
				view.Login(*params, errMap),
			),
			Status: http.StatusBadRequest,
		})
	}

	hash := u.HashedPassword.String
	match, err := argon2id.ComparePasswordAndHash(params.Password, hash)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login | Account",
				view.Error(
					"Failed to validate credentials",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}
	if !match {
		errMap["email"] = errors.New("invalid E-Mail or password")
		errMap["password"] = errMap["email"]
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login | Account",
				view.Login(*params, errMap),
			),
			Status: http.StatusBadRequest,
		})
	}

	sessionID, err := util.GenerateRandom(25)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login | Account",
				view.Error(
					"Failed to generate session id",
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

	expiresAt := time.Now().Add(time.Hour * 4)

	_, err = s.queries.CreateSession(
		c.Request().Context(),
		sqlc.CreateSessionParams{
			ID:        sessionID,
			UserID:    u.ID,
			ExpiresAt: expiresAt,
			Os:        sql.NullString{String: ua.OS, Valid: true},
			Browser:   browser,
		},
	)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login | Account",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	c.SetCookie(&http.Cookie{
		Name:     "auth_session",
		Value:    sessionID,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})

	returnTo := params.ReturnTo
	if returnTo == "" {
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
