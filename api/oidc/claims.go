package oidc

import "github.com/golang-jwt/jwt/v5"

type AccessTknClaims struct {
	jwt.RegisteredClaims
	UserID int64    `json:"user_id"`
	Scopes []string `json:"scopes"`
}

type IDTknClaims struct {
	jwt.RegisteredClaims
	AuthSessionID string `json:"auth_session_id"`
}
