package turnstile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const challengeURI = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

func VerifyCaptcha(ctx context.Context, r Request) (*Response, error) {
	req, err := newRequest(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("failed to create new http request: %w", err)
	}

	httpRes, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make http request: %w", err)
	}
	defer httpRes.Body.Close()

	if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status code: %s", httpRes.Status)
	}

	res := new(actualResponse)
	if err := json.NewDecoder(httpRes.Body).Decode(res); err != nil {
		return nil, fmt.Errorf("failed to decode turnstile response: %w", err)
	}

	return res.ToClientResponse()
}

func newRequest(ctx context.Context, r Request) (*http.Request, error) {
	q := make(url.Values)
	q.Set("secret", r.Secret)
	q.Set("response", r.Token)
	if r.IP != "" {
		q.Set("remoteip", r.IP)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		challengeURI,
		strings.NewReader(q.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}
