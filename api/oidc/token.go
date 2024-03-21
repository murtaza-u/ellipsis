package oidc

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/murtaza-u/account/api/util"
	"github.com/murtaza-u/account/internal/sqlc"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type tknParams struct {
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	Code         string `form:"code"`
	GrantType    string `form:"grant_type"`
}

type tknResp struct {
	Err       string `json:"error,omitempty"`
	ErrDesc   string `json:"error_description,omitempty"`
	AccessTkn string `json:"access_token,omitempty"`
	TknType   string `json:"token_type,omitempty"`
	ExpiresIn int    `json:"expires_in,omitempty"`
	Scope     string `json:"scope,omitempty"`
	IDTkn     string `json:"id_token,omitempty"`
}

func (a API) Token(c echo.Context) error {
	params := new(tknParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "bad_request",
			ErrDesc: "failed to parse form data",
		})
	}
	if params.GrantType != "authorization_code" {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "bad_request",
			ErrDesc: "invalid or unsupported grant_type",
		})
	}
	v := a.cache.Get(params.Code)
	if v == nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "bad_request",
			ErrDesc: "invalid or malformed authorization code",
		})
	}
	metadata, ok := v.(authorizeMetadata)
	if !ok {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "bad_request",
			ErrDesc: "invalid or malformed authorization code",
		})
	}
	if metadata.ClientID != params.ClientID {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "unauthorized",
			ErrDesc: "invalid client id or secret",
		})
	}
	client, err := a.db.GetClient(c.Request().Context(), metadata.ClientID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "unauthorized",
			ErrDesc: "invalid client id or secret",
		})
	}
	match, err := argon2id.ComparePasswordAndHash(params.ClientSecret, client.SecretHash)
	if err != nil || !match {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "unauthorized",
			ErrDesc: "invalid client id or secret",
		})
	}

	accessTkn := jwt.NewWithClaims(jwt.SigningMethodEdDSA, AccessTknClaims{
		UserID: metadata.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://localhost:3000/",
			Subject:   "http://localhost:3000/userinfo",
			Audience:  jwt.ClaimStrings{metadata.ClientID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 1800)),
		},
	})
	accessTknStr, err := accessTkn.SignedString(a.key.priv)
	if err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "internal_error",
			ErrDesc: "failed to generate access token",
		})
	}

	sessionID, err := util.GenerateRandom(25)
	if err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "internal_error",
			ErrDesc: "failed to generate auth session id",
		})
	}

	idTknExp := time.Now().Add(time.Second * time.Duration(client.TokenExpiration))
	idTkn := jwt.NewWithClaims(jwt.SigningMethodEdDSA, IDTknClaims{
		AuthSessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "http://localhost:3000/",
			Subject:   metadata.ClientID,
			Audience:  jwt.ClaimStrings{metadata.ClientID},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(idTknExp),
		},
	})
	idTknStr, err := idTkn.SignedString(a.key.priv)
	if err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "internal_error",
			ErrDesc: "failed to generate id token",
		})
	}

	var os, browser sql.NullString
	if metadata.OS != "" {
		os.String = metadata.OS
		os.Valid = true
	}
	if metadata.Browser != "" {
		browser.String = metadata.Browser
		browser.Valid = true
	}

	_, err = a.db.CreateSession(c.Request().Context(), sqlc.CreateSessionParams{
		ID:        sessionID,
		UserID:    metadata.UserID,
		ClientID:  sql.NullString{String: metadata.ClientID, Valid: true},
		ExpiresAt: idTknExp,
		Os:        os,
		Browser:   browser,
	})
	if err != nil {
		return c.JSON(http.StatusBadRequest, tknResp{
			Err:     "internal_error",
			ErrDesc: "database operation failed",
		})
	}

	// invalidate auth code
	a.cache.Delete(params.Code)

	return c.JSON(http.StatusOK, tknResp{
		AccessTkn: accessTknStr,
		TknType:   "Bearer",
		ExpiresIn: 1800,
		Scope:     strings.Join(metadata.Scopes, " "),
		IDTkn:     idTknStr,
	})
}
