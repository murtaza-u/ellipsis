package main

import (
	"log"
	"os"

	"github.com/murtaza-u/account/api"
)

func main() {
	user := os.Getenv("ACCOUNT_MYSQL_USER")
	if user == "" {
		log.Fatal("ACCOUNT_MYSQL_USER not set")
	}
	pass := os.Getenv("ACCOUNT_MYSQL_PASSWORD")
	if pass == "" {
		log.Fatal("ACCOUNT_MYSQL_PASS not set")
	}
	database := os.Getenv("ACCOUNT_MYSQL_DATABASE")
	if database == "" {
		log.Fatal("ACCOUNT_MYSQL_DATABASE not set")
	}

	s, err := api.New(api.Config{
		DatabaseUser:     user,
		DatabasePassword: pass,
		Database:         database,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}
}
