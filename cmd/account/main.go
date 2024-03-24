package main

import (
	"log"
	"os"

	"github.com/murtaza-u/ellipsis/api"
	"github.com/murtaza-u/ellipsis/internal/conf"
)

const defaultConfPath = "/etc/ellipsis/config.yaml"

func main() {
	path := os.Getenv("ELLIPSIS_CONFIG")
	if path == "" {
		path = defaultConfPath
	}

	c, err := conf.New(path)
	if err != nil {
		log.Fatal(err)
	}

	err = c.Validate()
	if err != nil {
		log.Fatalf("failed to validate config %s", err.Error())
	}

	s, err := api.New(*c)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Start()
	if err != nil {
		log.Fatal(err)
	}
}
