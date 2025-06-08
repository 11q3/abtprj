package handlers

import (
	"abtprj/internal/app"
	_ "database/sql"
	"log"
	"net/http"
	"strconv"
	"time"
)

type AdminPageData struct {
	Todos           []app.Task
	CurrentSession  string
	TotalSessionDur time.Duration
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
	case r.Method == http.MethodPost && r.URL.Path == "/admin/create-goal":
		h.createGoal(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/admin/get-work-status":
		h.getWorkingStatusForToday(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/start-work-session":
		h.startWorkSession(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/end-work-session":
		h.endWorkSession(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/login":
		h.handleLogin(w, r)

	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) renderAdminPage(w http.ResponseWriter) {
	todos, err := h.AppService.GetTodoTasks()

	if err != nil {
		log.Printf("get tasks query error: %v", err)
		http.Error(w, "repository error", http.StatusInternalServerError)
		return
	}

	today := time.Now().Format("2006-01-02")
	sessions, err := h.AppService.GetWorkSessionsForDate(today)
	if err != nil {
		log.Printf("renderAdminPage GetWorkSessionsForDate error: %v", err)
		http.Error(w, "failed to get sessions", http.StatusInternalServerError)
		return
	}

	var lastSession app.WorkSession
	if len(sessions) > 0 {
		lastSession = sessions[len(sessions)-1]
	}

	loc, locErr := time.LoadLocation("Europe/Moscow")
	if locErr != nil {
		loc = time.Local
	}

	var currentSession string
	var totalDur time.Duration

	for _, sess := range sessions {
		if sess.EndTime != nil {
			totalDur += sess.EndTime.Sub(sess.StartTime)
		}
	}

	if len(sessions) > 0 && lastSession.EndTime == nil {
		now := time.Now().In(loc)
		startFmt := lastSession.StartTime.In(loc).Format("15:04:05")
		endFmt := now.Format("15:04:05")
		ongoingDur := now.Sub(lastSession.StartTime.In(loc)).Truncate(time.Second)
		currentSession = startFmt + " - " + endFmt + " (" + ongoingDur.String() + ")"
		totalDur += ongoingDur
	}

	isWorking, err := h.AppService.IsWorking()
	if err != nil {
		log.Printf("worklog query error: %v", err)
	}

	data := AdminPageData{
		Todos:           todos,
		CurrentSession:  currentSession,
		TotalSessionDur: totalDur.Truncate(time.Second),
		IsWorking:       isWorking,
	}

	if err := h.Templates.ExecuteTemplate(w, "admin.html", data); err != nil {
		log.Printf("template exec error: %v", err)
	}

}

func (h *Handler) createGoal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "incorrect form values", http.StatusBadRequest)
		return
	}

	name := r.FormValue("goal_name")
	desc := r.FormValue("goal_description")
	dueStr := r.FormValue("goal_due")

	if name == "" || dueStr == "" {
		http.Error(w, "missing goal name or due date", http.StatusBadRequest)
		return
	}

	due, err := time.Parse("2006-01-02", dueStr)
	if err != nil {
		http.Error(w, "invalid due date", http.StatusBadRequest)
		return
	}

	goal := app.Goal{Name: name, Description: desc, DueAt: &due}
	if err := h.AppService.CreateGoal(goal); err != nil {
		log.Printf("createGoal CreateGoal error: %v", err)
		http.Error(w, "failed to create goal", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h *Handler) addTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "incorrect form values", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	description := r.FormValue("description")

	if err := h.AppService.AddTask(name, description); err != nil {
		log.Printf("addTask AddTask error: %v", err)
		http.Error(w, "failed to add a task", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h *Handler) completeTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "incorrect form values", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "task name is not present", http.StatusBadRequest)
		return
	}

	if err := h.AppService.CompleteTask(name); err != nil {
		log.Printf("completeTask CompleteTask error: %v", err)
		http.Error(w, "failed to complete a task", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h *Handler) getWorkingStatusForToday(w http.ResponseWriter, r *http.Request) {
	today := time.Now().Format("2006-01-02")
	sessions, err := h.AppService.GetWorkSessionsForDate(today)
	if err != nil {
		log.Printf("getWorkingStatusForToday GetWorkSessionsForDate error: %v", err)
		http.Error(w, "failed to get working status", http.StatusInternalServerError)
		return
	}

	isWorking := false
	for _, sess := range sessions {
		if sess.EndTime == nil {
			isWorking = true
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"working": ` + strconv.FormatBool(isWorking) + `}`))
}

func (h *Handler) startWorkSession(w http.ResponseWriter, r *http.Request) {
	if err := h.AppService.StartWorkSession(); err != nil {
		log.Printf("startWorkSession StartWorkSession error: %v", err)
		http.Error(w, "failed to start a session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) endWorkSession(w http.ResponseWriter, r *http.Request) {
	if err := h.AppService.EndWorkSession(); err != nil {
		log.Printf("endWorkSession EndWorkSession error: %v", err)
		http.Error(w, "failed to end a session", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректные данные формы", http.StatusBadRequest)
		return
	}
	login := r.FormValue("login")
	password := r.FormValue("password")

	if err := h.AppService.LoginAdmin(login, password); err != nil {
		log.Printf("handleLogin LoginAdmin error: %v", err)
		h.Templates.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Incorrect login or password",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "authenticated",
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}
