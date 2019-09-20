package main

import (
	"log"

	"github.com/ArnaudCalmettes/microsocial/actions"
)

func main() {
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
