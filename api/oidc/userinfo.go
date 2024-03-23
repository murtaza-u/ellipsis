package oidc

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type UserInfo struct {
	Err       string `json:"error,omitempty"`
	ErrDesc   string `json:"error_description,omitempty"`
	Email     string `json:"email,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

func (a API) UserInfo(c echo.Context) error {
	header := c.Request().Header.Get("Authorization")
	if header == "" {
		return c.JSON(http.StatusBadRequest, UserInfo{
			Err:     "bad_request",
			ErrDesc: "missing Authorization header in request",
		})
	}
	tknStr, err := tknFromHeader(header)
	if err != nil {
		return c.JSON(http.StatusBadRequest, UserInfo{
			Err:     "bad_request",
			ErrDesc: err.Error(),
		})
	}
	tkn, err := jwt.ParseWithClaims(
		tknStr,
		&AccessTknClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return *a.key.pub, nil
		},
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, UserInfo{
			Err:     "bad_request",
			ErrDesc: "invalid or expired access token",
		})
	}

	exp, err := tkn.Claims.GetExpirationTime()
	if err != nil {
		return c.JSON(http.StatusBadRequest, UserInfo{
			Err:     "bad_request",
			ErrDesc: "invalid or expired access token",
		})
	}
	if time.Until(exp.Time) <= 0 {
		return c.JSON(http.StatusBadRequest, UserInfo{
			Err:     "bad_request",
			ErrDesc: "invalid or expired access token",
		})
	}

	claims, ok := tkn.Claims.(*AccessTknClaims)
	if !ok {
		return c.JSON(http.StatusBadRequest, UserInfo{
			Err:     "bad_request",
			ErrDesc: "invalid or expired access token",
		})
	}

	var isAuthz bool
	for _, s := range claims.Scopes {
		if s == ScopeProfile {
			isAuthz = true
		}
	}
	if !isAuthz {
		return c.JSON(http.StatusBadRequest, UserInfo{
			Err:     "unauthorized",
			ErrDesc: "access token does not contain the required scope",
		})
	}

	u, err := a.DB.GetUser(c.Request().Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusBadRequest, UserInfo{
				Err:     "bad_request",
				ErrDesc: "invalid user id",
			})
		}
		return c.JSON(http.StatusInternalServerError, UserInfo{
			Err:     "internal",
			ErrDesc: "database operation failed",
		})
	}

	return c.JSON(http.StatusOK, UserInfo{
		Email:     u.Email,
		AvatarURL: u.AvatarUrl.String,
	})
}

func tknFromHeader(h string) (string, error) {
	parts := strings.Split(h, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header")
	}
	return parts[1], nil
}
