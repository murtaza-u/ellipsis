package console

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/murtaza-u/ellipsis/api/apierr"
	"github.com/murtaza-u/ellipsis/api/middleware"
	"github.com/murtaza-u/ellipsis/api/render"
	"github.com/murtaza-u/ellipsis/api/util"
	"github.com/murtaza-u/ellipsis/internal/sqlc"
	"github.com/murtaza-u/ellipsis/view"
	"github.com/murtaza-u/ellipsis/view/layout"
	"github.com/murtaza-u/ellipsis/view/partial/console"

	"github.com/a-h/templ"
	"github.com/alexedwards/argon2id"
	"github.com/labstack/echo/v4"
)

func (a API) appsPage(c echo.Context) error {
	clients, err := a.db.GetClients(c.Request().Context())
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to read clients from db: %w", err),
			layout.Base(
				"Console - Apps | Ellipsis",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
		)
	}

	var avatarURL string
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		avatarURL = ctx.AvatarURL
	}

	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Console - Apps | Ellipsis",
			view.Console("/console/app", avatarURL, console.Apps(clients)),
		),
	})
}

func (a API) appPage(c echo.Context) error {
	id := c.Param("id")
	if len(id) != 25 {
		return render.Do(render.Params{
			Ctx: c,
			Component: layout.Base(
				"Console - App | Ellipsis",
				view.Error(
					"Invalid app ID",
					http.StatusBadRequest,
				),
			),
			Status: http.StatusBadRequest,
		})
	}

	client, err := a.db.GetClient(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: layout.Base(
					"Console - App | Ellipsis",
					view.Error(
						"App not found",
						http.StatusNotFound,
					),
				),
				Status: http.StatusBadRequest,
			})
		}
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to read client from db: %w", err),
			layout.Base(
				"Console - App | Ellipsis",
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			),
		)
	}

	var avatarURL string
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		avatarURL = ctx.AvatarURL
	}

	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Console - Apps | Ellipsis",
			view.Console("/console/app", avatarURL, console.App(client)),
		),
	})
}

func (API) createAppPage(c echo.Context) error {
	var avatarURL string
	if ctx, ok := c.(middleware.CtxWithAuthInfo); ok {
		avatarURL = ctx.AvatarURL
	}

	return render.Do(render.Params{
		Ctx: c,
		Component: layout.Base(
			"Console - Create App | Ellipsis",
			view.Console("/console/app", avatarURL, console.AppCreate()),
		),
	})
}

func (a API) createApp(c echo.Context) error {
	params := new(console.AppParams)
	if err := c.Bind(params); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"Failed to parse form",
				http.StatusBadRequest,
			),
			Status: http.StatusBadRequest,
		})
	}
	v := newAppValidator(*params)
	params, errMap := v.Validate()
	if errMap["name"] == nil {
		_, err := a.db.GetClientByName(c.Request().Context(), params.Name)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return apierr.New(
				http.StatusInternalServerError,
				fmt.Errorf("faield to read client by name: %w", err),
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			)
		}
		if err == nil {
			errMap["name"] = errors.New("name already in use")
		}
	}
	if len(errMap) != 0 {
		return render.Do(render.Params{
			Ctx:       c,
			Component: console.AppUpdateForm(*params, false, errMap),
			Status:    http.StatusBadRequest,
		})
	}

	// generating client credentials
	id, err := util.GenerateRandom(25)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to generate random string: %w", err),
			view.Error(
				"Failed to generate client id",
				http.StatusInternalServerError,
			),
		)
	}

	secret, err := util.GenerateRandom(70)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to generate random string: %w", err),
			view.Error(
				"Failed to generate client secret",
				http.StatusInternalServerError,
			),
		)
	}

	hash, err := argon2id.CreateHash(secret, argon2id.DefaultParams)
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to create argon2id hash: %w", err),
			view.Error(
				"Failed to hash secret",
				http.StatusInternalServerError,
			),
		)
	}

	var pictureUrl sql.NullString
	if params.LogoURL != "" {
		pictureUrl.Valid = true
		pictureUrl.String = params.LogoURL
	}

	var backchannelLogoutURL sql.NullString
	if params.BackchannelLogoutURL != "" {
		backchannelLogoutURL.Valid = true
		backchannelLogoutURL.String = params.BackchannelLogoutURL
	}

	_, err = a.db.CreateClient(c.Request().Context(), sqlc.CreateClientParams{
		ID:                   id,
		SecretHash:           hash,
		Name:                 params.Name,
		PictureUrl:           pictureUrl,
		AuthCallbackUrls:     params.AuthCallbackURLs,
		LogoutCallbackUrls:   params.LogoutCallbackURLs,
		BackchannelLogoutUrl: backchannelLogoutURL,
		TokenExpiration:      int64(params.IDTokenExpiration),
	})
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to insert client into db: %w", err),
			view.Error(
				"Database operation failed",
				http.StatusInternalServerError,
			),
		)
	}

	return render.Do(render.Params{
		Ctx:       c,
		Component: console.AppCreateResult(params.Name, id, secret),
		Status:    http.StatusCreated,
	})
}

func (a API) updateApp(c echo.Context) error {
	params := new(console.AppParams)
	if err := c.Bind(params); err != nil {
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"Failed to parse form",
				http.StatusBadRequest,
			),
			Status: http.StatusBadRequest,
		})
	}

	if len(params.ID) != 25 {
		return render.Do(render.Params{
			Ctx: c,
			Component: view.Error(
				"Invalid app id",
				http.StatusBadRequest,
			),
			Status: http.StatusBadRequest,
		})
	}
	_, err := a.db.GetClient(c.Request().Context(), params.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: view.Error(
					"App not found",
					http.StatusNotFound,
				),
				Status: http.StatusNotFound,
			})
		}
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to read client from db: %w", err),
			view.Error(
				"Database operation failed",
				http.StatusInternalServerError,
			),
		)
	}

	v := newAppValidator(*params)
	params, errMap := v.Validate()
	if errMap["name"] == nil {
		_, err = a.db.GetClientByNameForUnmatchingID(
			c.Request().Context(),
			sqlc.GetClientByNameForUnmatchingIDParams{
				Name: params.Name,
				ID:   params.ID,
			},
		)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return apierr.New(
				http.StatusInternalServerError,
				fmt.Errorf("failed to read client from db: %w", err),
				view.Error(
					"Database operation failed",
					http.StatusInternalServerError,
				),
			)
		}
		if err == nil {
			errMap["name"] = errors.New("name already in use")
		}
	}
	if len(errMap) != 0 {
		return render.Do(render.Params{
			Ctx:       c,
			Component: console.AppUpdateForm(*params, false, errMap),
			Status:    http.StatusBadRequest,
		})
	}

	var pictureUrl sql.NullString
	if params.LogoURL != "" {
		pictureUrl.Valid = true
		pictureUrl.String = params.LogoURL
	}

	var backchannelLogoutURL sql.NullString
	if params.BackchannelLogoutURL != "" {
		backchannelLogoutURL.Valid = true
		backchannelLogoutURL.String = params.BackchannelLogoutURL
	}

	err = a.db.UpdateClient(c.Request().Context(), sqlc.UpdateClientParams{
		ID:                   params.ID,
		Name:                 params.Name,
		PictureUrl:           pictureUrl,
		AuthCallbackUrls:     params.AuthCallbackURLs,
		LogoutCallbackUrls:   params.LogoutCallbackURLs,
		BackchannelLogoutUrl: backchannelLogoutURL,
		TokenExpiration:      int64(params.IDTokenExpiration),
	})
	if err != nil {
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("failed to update client in db: %w", err),
			view.Error(
				"Database operation failed",
				http.StatusInternalServerError,
			),
		)
	}

	return render.Do(render.Params{
		Ctx:       c,
		Component: console.AppUpdateForm(*params, true, map[string]error{}),
		Status:    http.StatusCreated,
	})
}

func (a API) deleteApp(c echo.Context) error {
	id := c.Param("id")
	err := a.db.DeleteClient(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return render.Do(render.Params{
				Ctx: c,
				Component: view.Error(
					"App not found",
					http.StatusNotFound,
				),
				Status: http.StatusNotFound,
			})
		}
		return apierr.New(
			http.StatusInternalServerError,
			fmt.Errorf("faield to delete client from db: %w", err),
			view.Error(
				"Database operation failed",
				http.StatusInternalServerError,
			),
		)
	}

	// redirect to "/console/app"
	r := c.Response()
	r.Header().Set("HX-Redirect", "/console/app")

	// render empty template
	h := templ.Handler(view.Empty(), templ.WithStatus(http.StatusOK))
	return h.Component.Render(c.Request().Context(), r)
}
