package util

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/mileusna/useragent"
	pswdValidator "github.com/wagslane/go-password-validator"
)

const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-_"

func GenerateRandom(n int) (string, error) {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		ret[i] = chars[num.Int64()]
	}
	return string(ret), nil
}

func PEMToEd25519PrivKey(data []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("invalid PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key")
	}
	ed25519Priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("failed to parse private key")
	}
	return ed25519Priv, nil
}

func PEMToEd25519PubKey(data []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid PEM block")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key")
	}
	ed25519Pub, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to parse public key")
	}
	return ed25519Pub, nil
}

func BrowserFromUA(ua useragent.UserAgent) string {
	if ua.IsChrome() {
		return "Chrome"
	}
	if ua.IsEdge() {
		return "Edge"
	}
	if ua.IsFirefox() {
		return "Firefox"
	}
	if ua.IsInternetExplorer() {
		return "IE"
	}
	if ua.IsOpera() {
		return "Opera"
	}
	if ua.IsOperaMini() {
		return "Opera Mini"
	}
	return ""
}

func ValidatePassword(pswd string) error {
	if len(pswd) < 8 || len(pswd) > 70 {
		return errors.New("password must be between 8 and 70 characters")
	}
	err := pswdValidator.Validate(pswd, 60)
	if err != nil {
		return err
	}
	return nil
}

func ReadURL(ctx context.Context, url string) (io.ReadSeeker, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return bytes.NewReader(data), nil
}
