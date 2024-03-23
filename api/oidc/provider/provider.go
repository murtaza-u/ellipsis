package provider

import (
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

type Provider interface {
	Login(echo.Context) error
	Callback(echo.Context) error
}

type Credentials struct {
	ClientID     string
	ClientSecret string
}

type CallbackParams struct {
	Err     string `query:"error"`
	ErrDesc string `query:"error_description"`
	Code    string `query:"code"`
	State   string `query:"state"`
}

type TokenSource struct {
	tkn *oauth2.Token
}

func newTokenSource(tkn *oauth2.Token) oauth2.TokenSource {
	return &TokenSource{tkn: tkn}
}

func (s TokenSource) Token() (*oauth2.Token, error) {
	return s.tkn, nil
}
