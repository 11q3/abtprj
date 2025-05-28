package main

import (
	"abtprj/db"
	"database/sql"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	_ "github.com/lib/pq"
)

type App struct {
	db        *sql.DB
	templates *template.Template
}

func main() {
	connStr := "postgres://postgres:1121231@localhost:5432/abtprj?sslmode=disable"
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer dbConn.Close()

	templates, err := template.ParseGlob(filepath.Join("static", "*.html"))
	if err != nil {
		log.Fatalf("failed to parse templates: %v", err)
	}

	app := &App{db: dbConn, templates: templates}

	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, app.routes()))
}

func (app *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.mainHandler)
	mux.HandleFunc("/worklog/", app.workLogHandler)
	mux.HandleFunc("/admin/", app.adminHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return mux
}

func (app *App) mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join("static", "index.html"))
}

func (app *App) workLogHandler(w http.ResponseWriter, r *http.Request) {
	start, end, err := parseDateRange(r.URL.Query().Get("date"))
	if err != nil {
		http.Error(w, "invalid date format", http.StatusBadRequest)
		return
	}

	rows, err := app.db.Query(
		`SELECT name, description, status, done_at
         FROM tasks
         WHERE status = 'done' AND done_at >= $1 AND done_at < $2`,
		start, end,
	)
	if err != nil {
		log.Printf("worklog query error: %v", err)
		http.Error(w, "db query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []db.Task
	for rows.Next() {
		var t db.Task
		if err := rows.Scan(&t.Name, &t.Description, &t.Status, &t.DoneAt); err != nil {
			log.Printf("scan error: %v", err)
			continue
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		log.Printf("rows error: %v", err)
	}

	if err := app.templates.ExecuteTemplate(w, "worklog.html", struct{ Tasks []db.Task }{tasks}); err != nil {
		log.Printf("template exec error: %v", err)
	}
}

func parseDateRange(dateStr string) (time.Time, time.Time, error) {
	now := time.Now()
	if dateStr == "" {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, start.Add(24 * time.Hour), nil
	}
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return parsed, parsed.Add(24 * time.Hour), nil
}

// adminHandler routes admin-related actions based on method and path.
func (app *App) adminHandler(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/admin/":
		app.listTasks(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/add-task":
		app.addTask(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/admin/complete-task":
		app.completeTask(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (app *App) listTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := app.db.Query("SELECT name, description, status FROM tasks WHERE status = 'todo'")
	if err != nil {
		log.Printf("listTasks query error: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []db.Task
	for rows.Next() {
		var t db.Task
		if err := rows.Scan(&t.Name, &t.Description, &t.Status); err != nil {
			log.Printf("scan error: %v", err)
			continue
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		log.Printf("rows error: %v", err)
	}

	if err := app.templates.ExecuteTemplate(w, "admin.html", struct{ Tasks []db.Task }{tasks}); err != nil {
		log.Printf("template exec error: %v", err)
	}
}

func (app *App) addTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	descr := r.FormValue("description")
	if _, err := app.db.Exec(
		"INSERT INTO tasks(name, description, status) VALUES ($1, $2, 'todo')",
		name, descr,
	); err != nil {
		log.Printf("addTask exec error: %v", err)
		http.Error(w, "db insert error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (app *App) completeTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "missing task name", http.StatusBadRequest)
		return
	}
	if _, err := app.db.Exec(
		"UPDATE tasks SET status = 'done', done_at = NOW() WHERE name = $1",
		name,
	); err != nil {
		log.Printf("completeTask exec error: %v", err)
		http.Error(w, "db update error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}
