package handlers

import (
	_ "database/sql"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"abtprj/internal/repository"
)

type AdminPageData struct {
	Todos           []repository.Task
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
	case r.Method == http.MethodPost && r.URL.Path == "/admin/login":
		h.login(w, r)

	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) renderAdminPage(w http.ResponseWriter) {
	todos, err := repository.GetTodoTasks(h.DB)

	if err != nil {
		log.Printf("get tasks query error: %v", err)
		http.Error(w, "repository error", http.StatusInternalServerError)
		return
	}

	data := AdminPageData{
		Todos: todos,
	}

	if err := h.Templates.ExecuteTemplate(w, "admin.html", data); err != nil {
		log.Printf("template exec error: %v", err)
	}
}

func (h *Handler) initDefaultAdmin() error {
	login := os.Getenv("LOGIN")
	if login == "" {
		login = "admin"
	}

	password := os.Getenv("PASSWORD")
	if password == "" {
		password = "admin"
	}

	isExists, err := repository.CheckIfAdminExists(h.DB)
	if err != nil {
		return err
	}
	if isExists {
		log.Println("admin already exists")
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = repository.GenerateAdmin(h.DB, login, hash)
	if err != nil {
		log.Printf("generate admin error: %v", err)
		return err
	}
	return nil
}

func (h *Handler) addTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	descr := r.FormValue("description")
	if err := repository.AddTask(h.DB, name, descr); err != nil {
		log.Printf("addTask exec error: %v", err)
		http.Error(w, "repository insert error", http.StatusInternalServerError)
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
	if err := repository.CompleteTask(h.DB, name); err != nil {
		log.Printf("completeTask exec error: %v", err)
		http.Error(w, "repository update error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h *Handler) getWorkingStatusForToday(w http.ResponseWriter, r *http.Request) {
	isWorking := false

	today := time.Now().Format("2006-01-02")
	sessions, err := repository.GetWorkingSessionsForDay(h.DB, today)
	if err != nil {
		log.Println("get working status error,", err)
		http.Error(w, "repository error", http.StatusInternalServerError)
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
	err := repository.StartWorkSession(h.DB)
	if err != nil {
		log.Printf("startWorkSession exec error: %v", err)
		http.Error(w, "repository error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) endWorkSession(w http.ResponseWriter, r *http.Request) {
	err := repository.EndWorkSession(h.DB)
	if err != nil {
		log.Printf("endWorkSession exec error: %v", err)
		http.Error(w, "repository error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) bool {
	isAdmin, err := repository.CheckIfAdminExists(h.DB)
	if err != nil {
		log.Printf("checkIfAdminExists exec error: %v", err)
		http.Error(w, "repository error", http.StatusInternalServerError)
		return false
	}

	const sessionCookieName = "admin_session"

	return isAdmin
}
