package main

import (
	"log"
	"os"

	"github.com/owenrumney/schnappit/internal/app"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	application := app.New()

	if err := application.Run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
