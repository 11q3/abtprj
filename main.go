package main

import (
	"log"
	"net/http"

	"abtprj/internal/app"
)

func main() {
	app, err := app.New()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}
	defer app.DB.Close()

	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, app.Router))
}
