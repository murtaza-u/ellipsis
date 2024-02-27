package console

import (
	"errors"
	"net/url"
	"strings"

	"github.com/murtaza-u/account/view/partial/console"
)

type AppValidator struct {
	console.AppParams
}

func newAppValidator(p console.AppParams) AppValidator {
	return AppValidator{AppParams: p}
}

func (v AppValidator) Validate() (*console.AppParams, map[string]error) {
	errMap := make(map[string]error)
	if err := v.validateName(); err != nil {
		errMap["name"] = err
	}
	if err := v.validateLogoURL(); err != nil {
		errMap["logo_url"] = err
	}
	if err := v.validateCallbackURLs(); err != nil {
		errMap["callback_urls"] = err
	}
	if err := v.validateIDTokenExpiration(); err != nil {
		errMap["id_token_expiration"] = err
	}
	return &v.AppParams, errMap
}

func (v *AppValidator) validateName() error {
	v.Name = strings.TrimSpace(v.Name)
	if len(v.Name) < 2 || len(v.Name) > 50 {
		return errors.New("name must be between 2 and 50 characters")
	}
	return nil
}

func (v *AppValidator) validateLogoURL() error {
	if v.LogoURL == "" {
		return nil
	}

	v.LogoURL = strings.TrimSpace(v.LogoURL)
	v.LogoURL = strings.TrimSuffix(v.LogoURL, "/")

	if len(v.LogoURL) > 100 {
		return errors.New("url too long")
	}
	_, err := url.ParseRequestURI(v.LogoURL)
	if err != nil {
		return errors.New("invalid URL")
	}
	return nil
}

func (v *AppValidator) validateCallbackURLs() error {
	if v.CallbackURLs == "" {
		return errors.New("missing callback URL")
	}
	if len(v.CallbackURLs) > 1000 {
		return errors.New("value too long")
	}

	var callbacks []string

	for _, callback := range strings.Split(v.CallbackURLs, ",") {
		callback = strings.TrimSpace(callback)
		_, err := url.ParseRequestURI(callback)
		if err != nil {
			return errors.New("one or more invalid URL")
		}
		callbacks = append(callbacks, callback)
	}

	v.CallbackURLs = strings.Join(callbacks, ",")
	return nil
}

func (v AppValidator) validateIDTokenExpiration() error {
	if v.IDTokenExpiration < 300 || v.IDTokenExpiration > 86400 {
		return errors.New("id token expiration must be between 300s to 86400s")
	}
	return nil
}
