package main

import (
	"log"

	"github.com/GeraAnggaraPutra/go-backup/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
