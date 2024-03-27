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

const maxUploadSize = 1024 * 512

func (a API) Profile(c echo.Context) error {
	var avatarURL string
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		avatarURL = ctx.AvatarURL
	}
	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"My Account | Ellipsis",
			view.Me("/", avatarURL, me.Profile(avatarURL)),
		),
	})
}

func (a API) ChangeAvatar(c echo.Context) error {
	var userID, avatarURL string
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		avatarURL = ctx.AvatarURL
		userID = ctx.UserID
	}

	avatar, err := c.FormFile("avatar")
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"My Account | Ellipsis",
				view.Me(
					"/", avatarURL,
					view.Error(
						"failed to parse form file",
						http.StatusBadRequest,
					),
				),
			),
			Status: http.StatusBadRequest,
		})
	}
	if avatar.Size > maxUploadSize {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"My Account | Ellipsis",
				view.Me(
					"/", avatarURL,
					me.ChangeAvatar(avatarURL, map[string]error{
						"avatar": errors.New("file too large"),
					}),
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	file, err := avatar.Open()
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"My Account | Ellipsis",
				view.Me(
					"/", avatarURL,
					view.Error(
						"failed to open form file",
						http.StatusInternalServerError,
					),
				),
			),
			Status: http.StatusInternalServerError,
		})
	}
	defer file.Close()

	if err := a.fs.Put(userID, file); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"My Account | Ellipsis",
				view.Me(
					"/", avatarURL,
					view.Error(
						"failed to upload avatar to file storage",
						http.StatusInternalServerError,
					),
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	url, err := a.fs.GetURL(userID)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"My Account | Ellipsis",
				view.Me(
					"/", avatarURL,
					view.Error(
						"failed to fetch avatar url from file storage",
						http.StatusInternalServerError,
					),
				),
			),
			Status: http.StatusInternalServerError,
		})
	}

	err = a.db.UpdateUserAvatar(
		c.Request().Context(),
		sqlc.UpdateUserAvatarParams{
			ID:        userID,
			AvatarUrl: sql.NullString{String: url, Valid: true},
		},
	)
	if err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"My Account | Ellipsis",
				view.Me(
					"/", avatarURL,
					view.Error(
						"database operation failed",
						http.StatusInternalServerError,
					),
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
				"/", url,
				me.ChangeAvatar(url, map[string]error{}),
			),
		),
	})
}

func (a API) ChangePasswordPage(c echo.Context) error {
	var userID string
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
				"My Account - Change Password | Ellipsis",
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
			"My Account - Change Password | Ellipsis",
			view.Me(
				"/change-password",
				u.AvatarUrl.String,
				me.ChangePassword(
					me.ChangePasswordParams{HasPswd: u.HashedPassword.Valid},
					map[string]error{},
					false,
				),
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

	var userID string
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		userID = ctx.UserID
	}
	if userID == "" {
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
