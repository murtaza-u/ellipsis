package sdk

import "golang.org/x/oauth2"

type tknSrc struct {
	tkn *oauth2.Token
}

// NewTokenSource returns an oauth2 TokenSource.
func NewTokenSource(tkn *oauth2.Token) oauth2.TokenSource {
	return &tknSrc{tkn: tkn}
}

func (t tknSrc) Token() (*oauth2.Token, error) {
	return t.tkn, nil
}
