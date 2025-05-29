package handlers

import (
	"abtprj/internal/db"
	"abtprj/internal/utils"
	"log"
	"net/http"
	"time"
)

type WorklogPageData struct {
	Dones           []db.Task
	CurrentSession  string
	TotalSessionDur time.Duration
	IsWorking       bool
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

	dones, err := db.GetDoneTasks(h.DB, start, end)
	if err != nil {
		log.Printf("worklog query error: %v", err)
		http.Error(w, "db query error", http.StatusInternalServerError)
		return
	}

	workSessions, err := db.GetWorkingSessionsForDay(h.DB, date)
	if err != nil {
		log.Printf("working status query error: %v", err)
		http.Error(w, "get working status error", http.StatusInternalServerError)
		return
	}
	var last db.WorkSession
	if workSessions != nil {
		last = workSessions[len(workSessions)-1]
	}

	currentSession := last.StartTime.String() + " - " + time.Now().String()
	var totalDur time.Duration
	for _, d := range workSessions {
		if d.EndTime.Valid {
			totalDur += d.EndTime.Time.Sub(d.StartTime)
		}
	}

	loc, err := time.LoadLocation("Europe/Moscow")

	if !last.EndTime.Valid {
		a := time.Now().In(loc)
		b := last.StartTime
		c := a.Sub(b)
		totalDur += c
	}

	log.Printf("StartTime raw: %v | Location: %v", last.StartTime, last.StartTime.Location())
	log.Printf("Now: %v | Location: %v", time.Now(), time.Now().Location())
	log.Printf("StartTime.Before(Now): %v", last.StartTime.Before(time.Now()))
	log.Printf("time.Since(StartTime): %v", time.Since(last.StartTime))

	//totalSessionDur := time.Since(start).String()

	data := WorklogPageData{
		Dones:           dones,
		CurrentSession:  currentSession,
		TotalSessionDur: totalDur,
		IsWorking:       workSessions != nil,
	}

	if err := h.Templates.ExecuteTemplate(w, "worklog.html", data); err != nil {
		log.Printf("template exec error: %v", err)
	}
}
