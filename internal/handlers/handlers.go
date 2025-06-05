package handlers

import (
	"abtprj/internal/app"
	"database/sql"
	"net/http"
	"text/template"
)

type Handler struct {
	DB         *sql.DB
	Templates  *template.Template
	AppService app.AppService
}

func NewHandler(db *sql.DB, templates *template.Template, appService app.AppService) *Handler {
	return &Handler{DB: db, Templates: templates, AppService: appService}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.MainHandler)
	mux.HandleFunc("/worklog/", h.WorkLogHandler)
	mux.HandleFunc("/stats/", h.StatsHandler)
	mux.HandleFunc("/admin/", h.requireAdmin(h.AdminHandler))
	mux.HandleFunc("/login", h.HandleLogin)
}
