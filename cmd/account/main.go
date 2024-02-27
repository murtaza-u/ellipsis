package main

import (
	"log"

	"github.com/murtaza-u/account/api"
)

func main() {
	s := api.New()
	err := s.Start()
	if err != nil {
		log.Fatal(err)
	}
}
