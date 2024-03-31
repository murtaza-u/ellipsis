package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"
)

func init() {
	os.MkdirAll("keys", 0700)
}

func main() {
	path := filepath.Join("keys", "ed25519")
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatal(err)
	}

	// write private key to disk in PEM format
	b, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatal(err)
	}
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}
	err = os.WriteFile(path, pem.EncodeToMemory(block), 0600)
	if err != nil {
		log.Fatal(err)
	}

	// write public key to disk in PEM format
	b, err = x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		log.Fatal(err)
	}
	block = &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}
	err = os.WriteFile(path+".pub", pem.EncodeToMemory(block), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
