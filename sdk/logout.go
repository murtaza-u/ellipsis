package sdk

import (
	"fmt"
	"net/url"

	"github.com/coreos/go-oidc/v3/oidc"
)

// LogoutConfig contains configuration used to construct the
// front-channel logout endpoint.
type LogoutConfig struct {
	// IDTkn is the raw ID token returned from the token endpoint.
	IDTkn string
	// SID refers to the session ID field within the ID token claim.
	//
	// It is optional.
	SID string
	// RedirectURI refers to the URI to which the user needs to be
	// redirected after a successful logout. This URI needs to be
	// configured in the Ellipsis console.
	//
	// It is optional.
	RedirectURI string
	// ClientID is the application ID available in the Ellipsis console.
	//
	// It is optional.
	ClientID string
	// State is a random string that is returned as-is in the query
	// parameter of the redirect URI. It is used to prevent CSRF
	// attacks.
	//
	// It is optional.
	State string
}

// EndSessionEndpoint constructs the end session endpoint for
// front-channel logout. The client application should redirect the user
// to this endpoint for front-channel logout.
func EndSessionEndpoint(p *oidc.Provider, c LogoutConfig) (string, error) {
	claims := new(providerClaims)
	if err := p.Claims(claims); err != nil {
		return "", fmt.Errorf("failed to decode provider claims: %w", err)
	}
	if claims.EndSessionEndpoint == "" {
		return "", fmt.Errorf("missing end_session_endpoint in provider claims")
	}

	q := make(url.Values)

	if c.IDTkn == "" {
		return "", fmt.Errorf("missing id token in config")
	}
	q.Set("id_token_hint", c.IDTkn)

	if c.SID != "" {
		q.Set("logout_hint", c.SID)
	}

	if c.RedirectURI != "" {
		q.Set("post_logout_redirect_uri", c.RedirectURI)
	}

	if c.ClientID != "" {
		q.Set("client_id", c.ClientID)
	}

	if c.State != "" {
		q.Set("state", c.State)
	}

	return claims.EndSessionEndpoint + "?" + q.Encode(), nil
}
