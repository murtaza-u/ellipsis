package oidc

import "github.com/golang-jwt/jwt/v5"

type AccessTknClaims struct {
	jwt.RegisteredClaims
	UserID int64    `json:"user_id"`
	Scopes []string `json:"scopes"`
}

type IDTknClaims struct {
	jwt.RegisteredClaims
	SID string `json:"sid"`
}
