package handlers

import (
	"abtprj/internal/app"
	"log"
	"net/http"
	"time"
)

type WorklogPageData struct {
	Dones           []app.Task
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

	dones, err := h.AppService.GetTasksForDate(date)
	if err != nil {
		log.Printf("worklog query error: %v", err)
		http.Error(w, "repository query error", http.StatusInternalServerError)
		return
	}

	workSessions, err := h.AppService.GetWorkSessionsForDate(date)
	if err != nil {
		log.Printf("working status query error: %v", err)
		http.Error(w, "get working status error", http.StatusInternalServerError)
		return
	}

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.Local
	}

	var lastSession app.WorkSession
	if len(workSessions) > 0 {
		lastSession = workSessions[len(workSessions)-1]
	}

	var currentSession string
	var totalDur time.Duration

	for _, sess := range workSessions {
		if sess.EndTime != nil {
			totalDur += sess.EndTime.Sub(sess.StartTime)
		}
	}

	if len(workSessions) > 0 && lastSession.EndTime == nil {
		now := time.Now().In(loc)
		startFmt := lastSession.StartTime.In(loc).Format("15:04:05")
		endFmt := now.Format("15:04:05")
		currentSessionDuration := now.Sub(lastSession.StartTime.In(loc)).Truncate(time.Second)
		currentSession = startFmt + " - " + endFmt + " (" + currentSessionDuration.String() + ")"
		totalDur += currentSessionDuration
	}

	var isWorking bool
	isWorking, err = h.AppService.IsWorking()
	if err != nil {
		log.Printf("worklog query error: %v", err)
	}

	var sessionStrings []string
	for _, sess := range workSessions {
		start := sess.StartTime.In(loc).Format("15:04:05")
		var entry string

		if sess.EndTime != nil {
			end := sess.EndTime.In(loc).Format("15:04:05")
			dur := sess.EndTime.Sub(sess.StartTime).Truncate(time.Second)
			entry = start + " - " + end + " (" + dur.String() + ")"
		} else {
			entry = start + " - (ongoing)"
		}

		sessionStrings = append(sessionStrings, entry)
	}

	totalDur = totalDur.Truncate(time.Second)

	data := WorklogPageData{
		Dones:           dones,
		CurrentSession:  currentSession,
		TotalSessionDur: totalDur,
		IsWorking:       isWorking,
		AllSessions:     sessionStrings,
	}

	if err := h.Templates.ExecuteTemplate(w, "worklog.html", data); err != nil {
		log.Printf("template exec error: %v", err)
	}
}
