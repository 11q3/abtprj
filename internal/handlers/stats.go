package handlers

import (
	"abtprj/internal/repository"
	"log"
	"net/http"
	"time"
)

type MonthLabel struct {
	Name string // “Jun”, “Jul”, …
	Col  int    // which grid column (2–54) the label should start in
}

type DayTasksStat struct {
	Date  string // "2023-06-01"
	Count int    // how many tasks/sessions that day
	Level int    // 0–4 shade
	Row   int    // grid‐row (2–8): 2=Sunday, 3=Monday … 8=Saturday
	Col   int    // grid‐column (2–54): week index + 2
}

type DaySessionsStat struct {
	Date       string // "2023-06-01"
	SessionDur time.Duration
	Level      int // 0–4 shade
	Row        int // grid‐row (2–8): 2=Sunday, 3=Monday … 8=Saturday
	Col        int // grid‐column (2–54): week index + 2
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
	var emptyStatTasks = generateEmptyDayTasksStats()

	taskStats := h.populateDayStatsWithTasks(emptyTasks)
	sessionStats := h.populateDayStatsWithWorkSessions(emptyStatTasks)

	data := struct {
		TaskContributions    []DayTasksStat
		SessionContributions []DaySessionsStat
	}{
		taskStats,
		sessionStats,
	}

	if err := h.Templates.ExecuteTemplate(w, "stats.html", data); err != nil {
		log.Printf("template exec error: %v", err)
	}
}

func generateEmptyDayStats() []DayTasksStat {
	days := make([]DayTasksStat, 0, 52*7)
	for week := 1; week <= 52; week++ {
		for dow := 0; dow < 7; dow++ {
			days = append(days, DayTasksStat{"", 0, 0, dow + 2, week + 1})
		}
	}
	return days
}

func generateEmptyDayTasksStats() []DaySessionsStat { //TODO temp
	days := make([]DaySessionsStat, 0, 52*7)
	for week := 1; week <= 52; week++ {
		for dow := 0; dow < 7; dow++ {
			days = append(days, DaySessionsStat{"", time.Second * 0, 0, dow + 2, week + 1})
		}
	}
	return days
}

func (h *Handler) populateDayStatsWithTasks(emptyDayStats []DayTasksStat) []DayTasksStat {
	year := time.Now().Year()
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC)

	tasks, err := repository.GetDoneTasks(h.DB, start, end)
	if err != nil {
		return nil
	}

	stats := make([]DayTasksStat, len(emptyDayStats))
	copy(stats, emptyDayStats)

	for _, t := range tasks {
		if !t.DoneAt.Valid {
			continue
		}
		d := t.DoneAt.Time

		_, isoWeek := d.ISOWeek()
		weekIdx := isoWeek - 1

		dowIdx := (int(d.Weekday()) + 6) % 7

		idx := weekIdx*7 + dowIdx
		if idx < 0 || idx >= len(stats) {
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

func (h *Handler) populateDayStatsWithWorkSessions(emptyDayStats []DaySessionsStat) []DaySessionsStat {
	year := time.Now().Year()
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC)

	sessions, err := repository.GetWorkingSessions(h.DB, start, end)
	if err != nil {
		return nil
	}

	stats := make([]DaySessionsStat, len(emptyDayStats))
	copy(stats, emptyDayStats)

	for _, s := range sessions {
		if !s.EndTime.Valid {
			continue
		}
		d := s.EndTime.Time

		_, isoWeek := d.ISOWeek()
		weekIdx := isoWeek - 1

		// convert Go’s Sunday=0…Saturday=6 → Monday=0…Sunday=6
		dowIdx := (int(d.Weekday()) + 6) % 7

		idx := weekIdx*7 + dowIdx
		if idx < 0 || idx >= len(stats) {
			continue
		}

		stats[idx].Date = d.Format("2006-01-02")
		stats[idx].SessionDur += s.EndTime.Time.Sub(s.StartTime)
		switch {
		case stats[idx].SessionDur > 8*time.Hour:
			stats[idx].Level = 4
		case stats[idx].SessionDur > 4*time.Hour:
			stats[idx].Level = 3

		case stats[idx].SessionDur > 2*time.Hour:
			stats[idx].Level = 2

		case stats[idx].SessionDur > 1*time.Second:
			stats[idx].Level = 1

		default:
			stats[idx].Level = 0
		}
		if stats[idx].Level > 4 {
			stats[idx].Level = 4
		}
	}

	return stats
}
