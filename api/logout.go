package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (s Server) Logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:    "auth_session",
		Value:   "",
		Expires: time.Unix(0, 0),
	})

	cookie, err := c.Cookie("auth_session")
	if err != nil {
		return c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

	s.queries.DeleteSession(c.Request().Context(), cookie.Value)
	return c.Redirect(http.StatusTemporaryRedirect, "/login")
}
