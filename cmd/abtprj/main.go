package main

import (
	"abtprj/internal/handlers"
	"abtprj/internal/repository"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"abtprj/internal/app"
	_ "github.com/lib/pq"
)

func main() {
	connStr := "postgres://postgres:1121231@localhost:5432/abtprj?sslmode=disable"
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer dbConn.Close()

	if err := repository.InitDefaultAdmin(dbConn); err != nil {
		log.Fatalf("InitDefaultAdmin error: %v", err)
	}

	funcMap := template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
	}
	templates, err := template.New("").Funcs(funcMap).ParseGlob(filepath.Join("templates", "*.html"))
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	svc := app.NewDefaultAppService(dbConn)
	h := handlers.NewHandler(dbConn, templates, svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	addr := ":8080"
	log.Printf("Listening on %s…", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
