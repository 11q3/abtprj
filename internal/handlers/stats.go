package handlers

import (
	"abtprj/internal/app"
	"log"
	"net/http"
	"time"
)

func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/stats/":
		h.renderStatsPage(w)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) renderStatsPage(w http.ResponseWriter) {
	year := time.Now().Year()

	taskStats, err := h.AppService.GetDayTaskStats(year)
	if err != nil {
		log.Printf("could not get task stats: %v", err)
		http.Error(w, "failed to load stats", http.StatusInternalServerError)
		return
	}

	sessionStats, err := h.AppService.GetDaySessionStats(year)
	if err != nil {
		log.Printf("could not get session stats: %v", err)
		http.Error(w, "failed to load stats", http.StatusInternalServerError)
		return
	}

	data := struct {
		TaskContributions    []app.DayTasksStat
		SessionContributions []app.DaySessionsStat
	}{
		taskStats,
		sessionStats,
	}

	if err := h.Templates.ExecuteTemplate(w, "stats.html", data); err != nil {
		log.Printf("template exec error: %v", err)
	}
}
