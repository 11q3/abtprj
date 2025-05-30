package handlers

import (
	"log"
	"net/http"
)

type MonthLabel struct {
	Name string // “Jun”, “Jul”, …
	Col  int    // which grid column (2–54) the label should start in
}

type DayStat struct {
	Date  string // "2023-06-01"
	Count int    // how many tasks/sessions that day
	Level int    // 0–4 shade
	Row   int    // grid‐row (2–8): 2=Sunday, 3=Monday … 8=Saturday
	Col   int    // grid‐column (2–54): week index + 2
}

func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/stats/":
		h.renderStatsPage(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) renderStatsPage(w http.ResponseWriter, r *http.Request) {
	var a = DayStat{
		"2025-05-29",
		4,
		4,
		6,
		22,
	}
	var c = DayStat{
		"2025-05-30",
		1,
		1,
		7,
		22,
	}
	var d = DayStat{
		"2025-05-25",
		2,
		1,
		2,
		21,
	}
	var b = []DayStat{a, c, d, a, c, d, a, c, d, a, c, d, a, c, d, a, c, d}

	var j = MonthLabel{
		"May",
		22,
	}

	var g = []MonthLabel{j}

	data := struct {
		TaskMonths           []MonthLabel
		TaskContributions    []DayStat
		SessionMonths        []MonthLabel
		SessionContributions []DayStat
	}{
		g,
		b,
		g,
		b,
	}

	if err := h.Templates.ExecuteTemplate(w, "stats.html", data); err != nil {
		log.Printf("template exec error: %v", err)
	}
}
