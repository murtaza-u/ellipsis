package me

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/api/util"
	"github.com/murtaza-u/ellipsis/internal/sqlc"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"
	"github.com/murtaza-u/ellipsis/view/partial/me"

	"github.com/a-h/templ"
	"github.com/alexedwards/argon2id"
	"github.com/labstack/echo/v4"
)

func (a API) ProfilePage(c echo.Context) error {
	var userID int64
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		userID = ctx.UserID
	}
	u, err := a.db.GetUser(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r := c.Response()
			r.Header().Set("HX-Redirect", "/logout")

			// render empty template
			h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
			return h.Component.Render(c.Request().Context(), r)
		}
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"My Account | Ellipsis",
				view.Error(
					"database operation failed",
					http.StatusInternalServerError,
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"My Account | Ellipsis",
			view.Me(
				"/me",
				u.AvatarUrl.String,
				me.Profile(u.HashedPassword.Valid),
			),
		),
	})
}

func (a API) ChangePassword(c echo.Context) error {
	form := new(me.ChangePasswordParams)
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

	var userID int64
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		userID = ctx.UserID
	}
	if userID == 0 {
		r := c.Response()
		r.Header().Set("HX-Redirect", "/logout")

		// render empty template
		h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
		return h.Component.Render(c.Request().Context(), r)
	}

	u, err := a.db.GetUser(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r := c.Response()
			r.Header().Set("HX-Redirect", "/logout")

			// render empty template
			h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
			return h.Component.Render(c.Request().Context(), r)
		}
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"database operation failed",
				http.StatusInternalServerError,
			),
			Status: http.StatusInternalServerError,
		})
	}

	var match bool
	hash := u.HashedPassword
	if hash.Valid {
		match, err = argon2id.ComparePasswordAndHash(form.OldPassword, hash.String)
		if err != nil {
			return render.Do(render.Params{
				Ctx: c,
				Component: view.Error(
					"failed to validate old password",
					http.StatusInternalServerError,
				),
				Status: http.StatusInternalServerError,
			})
		}
	}

	errMap := make(map[string]error)
	if hash.Valid && !match {
		errMap["old_password"] = fmt.Errorf("incorrect password")
	}
	if err := util.ValidatePassword(form.NewPassword); err != nil {
		errMap["new_password"] = err
	}
	if form.NewPassword != form.NewConfirmPassword {
		errMap["new_password"] = fmt.Errorf("passwords do not match")
		errMap["new_confirm_password"] = errMap["new_password"]
	}

	form.HasPswd = hash.Valid

	if len(errMap) != 0 {
		return render.Do(render.Params{
			Ctx:       c,
			Component: me.ChangePassword(*form, errMap, false),
			Status:    http.StatusBadRequest,
		})
	}

	newHash, err := argon2id.CreateHash(form.NewPassword, argon2id.DefaultParams)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"failed to hash new password",
				http.StatusInternalServerError,
			),
			Status: http.StatusInternalServerError,
		})
	}

	err = a.db.UpdateUserPasswordHash(
		c.Request().Context(),
		sqlc.UpdateUserPasswordHashParams{
			ID:             u.ID,
			HashedPassword: sql.NullString{String: newHash, Valid: true},
		},
	)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"database operation failed",
				http.StatusInternalServerError,
			),
			Status: http.StatusInternalServerError,
		})
	}

	return render.Do(render.Params{
		Ctx: c,
		Component: me.ChangePassword(
			me.ChangePasswordParams{HasPswd: true},
			map[string]error{},
			true,
		),
	})
}
