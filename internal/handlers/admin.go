package handlers

import (
	"log"
	"net/http"

	"abtprj/internal/db"
)

func (h *Handler) AdminHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/admin/":
		h.listTasks(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/add-task":
		h.addTask(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/complete-task":
		h.completeTask(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := db.GetTodoTasks(h.DB)
	if err != nil {
		log.Printf("listTasks query error: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	if err := h.Templates.ExecuteTemplate(w, "admin.html", struct{ Tasks []db.Task }{tasks}); err != nil {
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
