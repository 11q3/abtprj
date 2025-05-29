package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"abtprj/internal/db"
)

type AdminPageData struct {
	Todos           []db.Task
	CurrentSession  string
	TotalSessionDur string
	IsWorking       bool
}

func (h *Handler) AdminHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/admin/":
		h.renderAdminPage(w)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/add-task":
		h.addTask(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/complete-task":
		h.completeTask(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/admin/get-work-status":
		h.getWorkingStatusForToday(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/start-work-session":
		h.startWorkSession(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/end-work-session":
		h.endWorkSession(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) renderAdminPage(w http.ResponseWriter) {
	todos, err := db.GetTodoTasks(h.DB)

	if err != nil {
		log.Printf("get tasks query error: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	data := AdminPageData{
		Todos: todos,
		//CurrentSession:   sessions,
		//TotalSessionDur: total.String(),
		//IsWorking:       isWorking,
	}

	if err := h.Templates.ExecuteTemplate(w, "admin.html", data); err != nil {
		log.Printf("template exec error: %v", err)
	}
}

func (h *Handler) addTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	descr := r.FormValue("description")
	if err := db.AddTask(h.DB, name, descr); err != nil {
		log.Printf("addTask exec error: %v", err)
		http.Error(w, "db insert error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h *Handler) completeTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "missing task name", http.StatusBadRequest)
		return
	}
	if err := db.CompleteTask(h.DB, name); err != nil {
		log.Printf("completeTask exec error: %v", err)
		http.Error(w, "db update error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h *Handler) getWorkingStatusForToday(w http.ResponseWriter, r *http.Request) {
	isWorking := false

	today := time.Now().Format("2006-01-02")
	sessions, err := db.GetWorkingSessionsForDay(h.DB, today)
	if err != nil {
		log.Println("get working status error,", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	for _, s := range sessions {
		if !s.EndTime.Valid {
			isWorking = true
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"working": ` + strconv.FormatBool(isWorking) + `}`))
}

func (h *Handler) startWorkSession(w http.ResponseWriter, r *http.Request) {
	err := db.StartWorkSession(h.DB)
	if err != nil {
		log.Printf("startWorkSession exec error: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) endWorkSession(w http.ResponseWriter, r *http.Request) {
	err := db.EndWorkSession(h.DB)
	if err != nil {
		log.Printf("endWorkSession exec error: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
}
