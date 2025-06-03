package app

import (
	"abtprj/internal/db"
	"abtprj/internal/handlers"
	"database/sql"
	_ "github.com/lib/pq"
	"net/http"
	"path/filepath"
	"text/template"
)

type App struct {
	DB        *sql.DB
	Templates *template.Template
	Router    http.Handler
}

func New() (*App, error) {
	connStr := "postgres://postgres:1121231@localhost:5432/abtprj?sslmode=disable"
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.InitDefaultAdmin(dbConn); err != nil {
		return nil, err
	}

	templates, err := template.ParseGlob(filepath.Join("static", "*.html"))
	if err != nil {
		return nil, err
	}

	h := handlers.NewHandler(dbConn, templates)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return &App{DB: dbConn, Templates: templates, Router: mux}, nil
}
