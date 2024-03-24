package oidc

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type config struct {
	Issuer                            string   `json:"issuer"`
	AuthzEndp                         string   `json:"authorization_endpoint"`
	TknEndp                           string   `json:"token_endpoint"`
	UserinfoEndp                      string   `json:"userinfo_endpoint"`
	JWKsURI                           string   `json:"jwks_uri"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ResponseModesSupported            []string `json:"response_modes_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTknSigningAlgValuesSupported    []string `json:"id_token_signing_alg_values_supported"`
	TknEndpAuthMethodsSupported       []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	RequestURIParamSupported          bool     `json:"request_uri_parameter_supported"`
	RequestParamSupported             bool     `json:"request_parameter_supported"`
	EndSessionEndpoint                string   `json:"end_session_endpoint"`
	BackchannelLogoutSupported        bool     `json:"backchannel_logout_supported"`
	BackchannelLogoutSessionSupported bool     `json:"backchannel_logout_session_supported"`
}

func (API) configuration(c echo.Context) error {
	return c.JSON(http.StatusOK, config{
		Issuer:                         "http://localhost:3000/",
		AuthzEndp:                      "http://localhost:3000/authorize",
		TknEndp:                        "http://localhost:3000/oauth/token",
		UserinfoEndp:                   "http://localhost:3000/userinfo",
		JWKsURI:                        "http://localhost:3000/.well-known/jwks.json",
		ScopesSupported:                []string{"openid", "profile"},
		ResponseTypesSupported:         []string{"code"},
		ResponseModesSupported:         []string{"query"},
		SubjectTypesSupported:          []string{"public"},
		IDTknSigningAlgValuesSupported: []string{"EdDSA"},
		TknEndpAuthMethodsSupported:    []string{"client_secret_post"},
		ClaimsSupported: []string{
			"iss",
			"aud",
			"sub",
			"iat",
			"exp",
			"sid",
		},
		RequestURIParamSupported:          false,
		RequestParamSupported:             false,
		EndSessionEndpoint:                "http://localhost:3000/oidc/logout",
		BackchannelLogoutSupported:        false,
		BackchannelLogoutSessionSupported: false,
	})
}
