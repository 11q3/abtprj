package handlers

import (
	"database/sql"
	"net/http"
	"text/template"
)

type Handler struct {
	DB        *sql.DB
	Templates *template.Template
}

func NewHandler(db *sql.DB, templates *template.Template) *Handler {
	return &Handler{DB: db, Templates: templates}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.MainHandler)
	mux.HandleFunc("/worklog/", h.WorkLogHandler)
	mux.HandleFunc("/stats/", h.StatsHandler)
	mux.HandleFunc("/admin/", h.AdminHandler)
}
