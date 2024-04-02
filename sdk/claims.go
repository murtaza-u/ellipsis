package sdk

import "github.com/golang-jwt/jwt/v5"

// IDTokenClaims represents the claims contained within the ID token
// issued by Ellipsis.
type IDTokenClaims struct {
	jwt.RegisteredClaims
	SID string `json:"sid"`
}

// UserInfoClaims represents the user's information returned from
// Ellipsis' user info endpoint.
type UserInfoClaims struct {
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// LogoutTokenClaims represents the claims within the logout token
// issued by Ellipsis during back-channel logout.
type LogoutTokenClaims struct {
	jwt.RegisteredClaims
	Events map[string]struct{} `json:"events"`
	SID    string              `json:"sid"`
}

type providerClaims struct {
	EndSessionEndpoint string `json:"end_session_endpoint"`
}
