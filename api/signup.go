package api

import (
	"database/sql"
	"errors"
	"net/http"
	"net/mail"

	"github.com/murtaza-u/account/api/render"
	"github.com/murtaza-u/account/internal/sqlc"
	"github.com/murtaza-u/account/view"
	"github.com/murtaza-u/account/view/layout"

	"github.com/a-h/templ"
	"github.com/alexedwards/argon2id"
	"github.com/labstack/echo/v4"
	pswdValidator "github.com/wagslane/go-password-validator"
)

func (Server) SignUpPage(c echo.Context) error {
	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Sign Up | Account",
			view.SignUp(view.SignUpParams{
				ReturnTo: c.QueryParam("return_to"),
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
				"Sign Up | Account",
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
	if err := validatePassword(params.Password); err != nil {
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
				"Sign Up | Account",
				view.SignUp(*params, errMap),
			),
			Status: http.StatusBadRequest,
		})
	}

	hash, err := argon2id.CreateHash(params.Password, argon2id.DefaultParams)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Sign Up | Account",
				view.Error(
					"Failed to hash password",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	_, err = s.queries.CreateUser(c.Request().Context(), sqlc.CreateUserParams{
		Email:          params.Email,
		HashedPassword: sql.NullString{String: hash, Valid: true},
	})
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Sign Up | Account",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
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

func validatePassword(pswd string) error {
	if len(pswd) < 8 || len(pswd) > 70 {
		return errors.New("password must be between 8 and 70 characters")
	}
	err := pswdValidator.Validate(pswd, 60)
	if err != nil {
		return err
	}
	return nil
}
