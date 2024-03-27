package oidc

import "github.com/golang-jwt/jwt/v5"

type AccessTknClaims struct {
	jwt.RegisteredClaims
	UserID string   `json:"user_id"`
	Scopes []string `json:"scopes"`
}

type IDTknClaims struct {
	jwt.RegisteredClaims
	SID string `json:"sid"`
}

type LogoutTknClaims struct {
	jwt.RegisteredClaims
	Events map[string]struct{} `json:"events"`
	SID    string              `json:"sid"`
}
