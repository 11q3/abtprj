package handlers

import (
	"abtprj/internal/repository"
	"abtprj/internal/utils"
	"log"
	"net/http"
	"time"
)

type WorklogPageData struct {
	Dones           []repository.Task
	CurrentSession  string
	TotalSessionDur time.Duration
	IsWorking       bool
	AllSessions     []string
}

func (h *Handler) WorkLogHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/worklog/":
		h.renderWorklogPage(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) renderWorklogPage(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	start, end, err := utils.ParseDateRange(date)
	if err != nil {
		http.Error(w, "invalid date format", http.StatusBadRequest)
		return
	}

	dones, err := repository.GetDoneTasks(h.DB, start, end)
	if err != nil {
		log.Printf("worklog query error: %v", err)
		http.Error(w, "repository query error", http.StatusInternalServerError)
		return
	}

	workSessions, err := repository.GetWorkingSessionsForDay(h.DB, date)
	if err != nil {
		log.Printf("working status query error: %v", err)
		http.Error(w, "get working status error", http.StatusInternalServerError)
		return
	}

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.Local
	}

	var last repository.WorkSession
	if len(workSessions) > 0 {
		last = workSessions[len(workSessions)-1]
	}

	var currentSession string
	var totalDur time.Duration
	for _, sess := range workSessions {
		if sess.EndTime.Valid {
			totalDur += sess.EndTime.Time.Sub(sess.StartTime)
		}
	}

	if len(workSessions) > 0 && !last.EndTime.Valid {
		now := time.Now().In(loc)
		startFmt := last.StartTime.In(loc).Format("15:04:05")
		endFmt := now.Format("15:04:05")
		currentSessionDuration := time.Since(last.StartTime)
		currentSession = startFmt + " - " + endFmt + " (" + currentSessionDuration.Truncate(time.Second).String() + ")"
		totalDur += now.Sub(last.StartTime.In(loc))
	}

	var sessionStrings []string
	for _, sess := range workSessions {
		start := sess.StartTime.In(loc).Format("15:04:05")
		var entry string

		if sess.EndTime.Valid {
			end := sess.EndTime.Time.In(loc).Format("15:04:05")
			dur := sess.EndTime.Time.Sub(sess.StartTime).Truncate(time.Second)
			entry = start + " - " + end + " (" + dur.String() + ")"
			totalDur += dur
		} else {
			entry = start + " - (ongoing)"
			now := time.Now().In(loc)
			totalDur += now.Sub(sess.StartTime.In(loc)).Truncate(time.Second)
		}

		sessionStrings = append(sessionStrings, entry)
	}

	totalDur = totalDur.Truncate(time.Second)

	data := WorklogPageData{
		Dones:           dones,
		CurrentSession:  currentSession,
		TotalSessionDur: totalDur,
		IsWorking:       len(workSessions) > 0 && !last.EndTime.Valid,
		AllSessions:     sessionStrings,
	}

	if err := h.Templates.ExecuteTemplate(w, "worklog.html", data); err != nil {
		log.Printf("template exec error: %v", err)
	}
}
