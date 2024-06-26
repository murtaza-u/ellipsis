package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/murtaza-u/ellipsis/api/apierr"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/api/util"
	"github.com/murtaza-u/ellipsis/internal/sqlc"
	"github.com/murtaza-u/ellipsis/turnstile"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"

	"github.com/a-h/templ"
	"github.com/alexedwards/argon2id"
	"github.com/labstack/echo/v4"
)

func (s Server) SignUpPage(c echo.Context) error {
	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Sign Up | Ellipsis",
			view.SignUp(view.SignUpParams{
				ReturnTo:  c.QueryParam("return_to"),
				Providers: s.Providers,
				Captcha:   s.Captcha,
			}, map[string]error{}),
		),
	})
}

func (s Server) SignUp(c echo.Context) error {
	params := new(view.SignUpParams)
	if err := c.Bind(params); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Sign Up | Ellipsis",
				view.Error("Failed to parse form", http.StatusBadRequest),
			),
			Status: http.StatusBadRequest,
		})
	}
	params.ReturnTo = c.QueryParam("return_to")
	params.Providers = s.Providers
	params.Captcha = s.Captcha

	if s.Captcha.Turnstile.Enable {
		err := s.verifyTurnstileCaptcha(c, params.TurnstileToken)
		if err != nil {
			return err
		}
	}

	errMap := make(map[string]error)

	if err := validateEmail(params.Email); err != nil {
		errMap["email"] = err
	}
	exists, err := s.userExists(c.Request().Context(), params.Email)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to read user from db: %w", err),
			layout.Base(
				"Sign Up | Ellipsis",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
		)
	}
	if exists {
		errMap["email"] = errors.New("user already exists")
	}

	if err := util.ValidatePassword(params.Password); err != nil {
		errMap["password"] = err
	}
	if params.Password != params.ConfirmPassword {
		errMap["password"] = errors.New("passwords do not match")
		errMap["confirm_password"] = errMap["password"]
	}

	if len(errMap) != 0 {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Sign Up | Ellipsis",
				view.SignUp(*params, errMap),
			),
			Status: http.StatusBadRequest,
		})
	}

	hash, err := argon2id.CreateHash(params.Password, argon2id.DefaultParams)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to create argon2id hash: %w", err),
			layout.Base(
				"Sign Up | Ellipsis",
				view.Error(
					"Failed to hash password",
					http.StatusInternalServerError,
				),
			),
		)
	}

	userID, err := util.GenerateRandom(25)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to generate random string: %w", err),
			layout.Base(
				"Sign Up | Ellipsis",
				view.Error(
					"failed to generate user id",
					http.StatusInternalServerError,
				),
			),
		)
	}

	_, err = s.queries.CreateUser(c.Request().Context(), sqlc.CreateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: sql.NullString{String: hash, Valid: true},
	})
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to insert new user in db: %w", err),
			layout.Base(
				"Sign Up | Ellipsis",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
		)
	}

	returnTo := params.ReturnTo
	if returnTo == "" {
		returnTo = "/login"
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

func (s Server) userExists(ctx context.Context, email string) (bool, error) {
	_, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func (s Server) verifyTurnstileCaptcha(c echo.Context, tkn string) error {
	if tkn == "" {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Sign Up | Ellipsis",
				view.Error("Failed to verify captcha", http.StatusBadRequest),
			),
			Status: http.StatusBadRequest,
		})
	}

	res, err := turnstile.VerifyCaptcha(
		c.Request().Context(),
		turnstile.Request{
			Secret: s.Captcha.Turnstile.SecretKey,
			Token:  tkn,
			IP:     c.RealIP(),
		},
	)
	if err != nil {
		return apierr.New(
			http.StatusBadRequest,
			fmt.Errorf("failed to verify captcha: %w", err),
			layout.Base(
				"Sign Up | Ellipsis",
				view.Error("Failed to verify captcha", http.StatusBadRequest),
			),
		)
	}

	if res.Success {
		return nil
	}

	return apierr.New(
		http.StatusBadRequest,
		fmt.Errorf("failed to verify captcha: %w", errors.Join(res.Errors...)),
		layout.Base(
			"Sign Up | Ellipsis",
			view.Error("Failed to verify captcha", http.StatusBadRequest),
		),
	)
}

func validateEmail(email string) error {
	if len(email) > 50 {
		return errors.New("too long")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid E-Mail")
	}
	return nil
}
