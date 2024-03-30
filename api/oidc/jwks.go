package oidc

import (
	"encoding/base64"
	"net/http"

	"github.com/labstack/echo/v4"
)

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Crv string `json:"crv"`
	X   string `json:"x"`
}

type Keys struct {
	JWKs []JWK `json:"keys"`
}

func (a API) JWKs(c echo.Context) error {
	return c.JSON(http.StatusOK, Keys{
		JWKs: []JWK{
			{
				Kty: "OKP",
				Kid: "ed25519-key-1",
				Alg: "EdDSA",
				Use: "sig",
				Crv: "Ed25519",
				X:   base64.RawStdEncoding.EncodeToString(a.Key.Pub),
			},
		},
	})
}
