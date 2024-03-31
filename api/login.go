package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/murtaza-u/ellipsis/api/apierr"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/api/util"
	"github.com/murtaza-u/ellipsis/internal/sqlc"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"

	"github.com/a-h/templ"
	"github.com/alexedwards/argon2id"
	"github.com/labstack/echo/v4"
)

func (s Server) LoginPage(c echo.Context) error {
	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Login | Ellipsis",
			view.Login(view.LoginParams{
				ReturnTo:  c.QueryParam("return_to"),
				Providers: s.Providers,
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
				"Login | Ellipsis",
				view.Error("Failed to parse form", http.StatusBadRequest),
			),
			Status: http.StatusBadRequest,
		})
	}
	params.ReturnTo = c.QueryParam("return_to")
	params.Providers = s.Providers

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
				"Login | Ellipsis",
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
					"Login | Ellipsis",
					view.Login(*params, errMap),
				),
				Status: http.StatusBadRequest,
			})
		}
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to read user from db: %w", err),
			layout.Base(
				"Login | Ellipsis",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
		)
	}
	if !u.HashedPassword.Valid {
		errMap["email"] = errors.New("invalid E-Mail or password")
		errMap["password"] = errMap["email"]
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login | Ellipsis",
				view.Login(*params, errMap),
			),
			Status: http.StatusBadRequest,
		})
	}

	hash := u.HashedPassword.String
	match, err := argon2id.ComparePasswordAndHash(params.Password, hash)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to compare argon2id hash with plain-text: %w", err),
			layout.Base(
				"Login | Ellipsis",
				view.Error(
					"Failed to validate credentials",
					http.StatusInternalServerError,
				),
			),
		)
	}
	if !match {
		errMap["email"] = errors.New("invalid E-Mail or password")
		errMap["password"] = errMap["email"]
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Login | Ellipsis",
				view.Login(*params, errMap),
			),
			Status: http.StatusBadRequest,
		})
	}

	sessionID, err := util.GenerateRandom(25)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to generate random string: %w", err),
			layout.Base(
				"Login | Ellipsis",
				view.Error(
					"Failed to generate session id",
					http.StatusInternalServerError,
				),
			),
		)
	}

	fingerprint := util.ParseUA(c.Request().Header.Get("User-Agent"))
	expiresAt := time.Now().Add(time.Hour * 4)

	_, err = s.queries.CreateSession(
		c.Request().Context(),
		sqlc.CreateSessionParams{
			ID:        sessionID,
			UserID:    u.ID,
			ExpiresAt: expiresAt,
			Browser:   fingerprint.Browser,
			Os:        fingerprint.OS,
		},
	)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to insert session into db: %w", err),
			layout.Base(
				"Login | Ellipsis",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
		)
	}

	c.SetCookie(&http.Cookie{
		Name:     "auth_session",
		Value:    sessionID,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
		Path:     "/",
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
