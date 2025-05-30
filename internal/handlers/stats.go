package handlers

import (
	"abtprj/internal/db"
	"log"
	"net/http"
	"time"
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
	var emptyTasks = generateEmptyDayStats()

	taskStats := h.populateDayStatsWithTasks(emptyTasks)
	sessionStats := h.populateDayStatsWithWorkSessions(emptyTasks)

	data := struct {
		TaskContributions    []DayStat
		SessionContributions []DayStat
	}{
		taskStats,
		sessionStats,
	}

	if err := h.Templates.ExecuteTemplate(w, "stats.html", data); err != nil {
		log.Printf("template exec error: %v", err)
	}
}

func generateEmptyDayStats() []DayStat {
	days := make([]DayStat, 0, 52*7)
	for week := 1; week <= 52; week++ {
		for dow := 0; dow < 7; dow++ {
			days = append(days, DayStat{"", 0, 0, dow + 2, week + 1})
		}
	}
	return days
}

func (h *Handler) populateDayStatsWithTasks(emptyDayStats []DayStat) []DayStat {
	year := time.Now().Year()
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC)

	// fetch both the slice of tasks and error
	tasks, err := db.GetDoneTasks(h.DB, start, end)
	if err != nil {
		return nil
	}

	// copy your 52×7 grid
	stats := make([]DayStat, len(emptyDayStats))
	copy(stats, emptyDayStats)

	for _, t := range tasks {
		// unwrap the NullTime
		if !t.DoneAt.Valid {
			continue
		}
		d := t.DoneAt.Time

		// get ISO week (Mon=1…Sun=7 → week 1…53)
		_, isoWeek := d.ISOWeek()
		weekIdx := isoWeek - 1 // 0-based index

		// convert Go’s Sunday=0…Saturday=6 → Monday=0…Sunday=6
		dowIdx := (int(d.Weekday()) + 6) % 7

		idx := weekIdx*7 + dowIdx
		if idx < 0 || idx >= len(stats) {
			// skip out-of-range (e.g. ISO week 53 spill)
			continue
		}

		stats[idx].Date = d.Format("2006-01-02")
		stats[idx].Count++
		stats[idx].Level = stats[idx].Count
		if stats[idx].Level > 4 {
			stats[idx].Level = 4
		}
	}

	return stats
}

func (h *Handler) populateDayStatsWithWorkSessions(emptyDayStats []DayStat) []DayStat { //TODO later do with time
	year := time.Now().Year()
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC)

	// fetch both the slice of tasks and error
	tasks, err := db.GetWorkingSessions(h.DB, start, end)
	if err != nil {
		return nil
	}

	// copy your 52×7 grid
	stats := make([]DayStat, len(emptyDayStats))
	copy(stats, emptyDayStats)

	for _, s := range tasks {
		// unwrap the NullTime
		if !s.EndTime.Valid {
			continue
		}
		d := s.EndTime.Time

		// get ISO week (Mon=1…Sun=7 → week 1…53)
		_, isoWeek := d.ISOWeek()
		weekIdx := isoWeek - 1 // 0-based index

		// convert Go’s Sunday=0…Saturday=6 → Monday=0…Sunday=6
		dowIdx := (int(d.Weekday()) + 6) % 7

		idx := weekIdx*7 + dowIdx
		if idx < 0 || idx >= len(stats) {
			// skip out-of-range (e.g. ISO week 53 spill)
			continue
		}

		stats[idx].Date = d.Format("2006-01-02")
		stats[idx].Count++
		stats[idx].Level = stats[idx].Count
		if stats[idx].Level > 4 {
			stats[idx].Level = 4
		}
	}

	return stats
}
