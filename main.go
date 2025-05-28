package main

import (
	"abtprj/db"
	"database/sql"
	"log"
	"net/http"
	"text/template"
	"time"

	_ "github.com/lib/pq"
)

const addr = "http://localhost:8080"
const connStr = "postgres://postgres:1121231@localhost:5432/abtprj?sslmode=disable"

func main() {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("error while connecting to db")
		return
	}

	mainHandler := makeMainHandler()
	secondHandler := makeAdminHandler(db)
	workLogHandler := makeWorkLogHandler(db)

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/worklog/", workLogHandler)
	http.HandleFunc("/admin/", secondHandler)

	log.Printf("Starting server at %s", addr)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Fatal(
		http.ListenAndServe(":8080", nil))
}

func makeMainHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./static/index.html")
	}
}

func makeWorkLogHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("static/worklog.html")
		if err != nil {
			http.Error(w, "Template parsing error", 500)
			log.Println(err)
			return
		}

		rows, err := dbConn.Query("" +
			"SELECT id, date, start_time, end_time, duration FROM work_sessions ORDER BY date DESC")
		if err != nil {
			http.Error(w, "DB error", 500)
			log.Println(err)
			return
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				log.Println("rows closing error:", err)
				return
			}
		}(rows)

		var sessions []db.WorkSession

		for rows.Next() {
			var s db.WorkSession
			var id int
			var rawDate time.Time

			err := rows.Scan(&id, &rawDate, &s.StartTime, &s.EndTime, &s.Duration)
			if err != nil {
				log.Println("Scan error:", err)
				continue
			}

			s.Date = rawDate.Format("2006-01-02")

			taskRows, err := dbConn.Query(
				"SELECT name, description, status, created_at, done_at FROM tasks WHERE session_id = $1", id)

			if err != nil {
				log.Println("Task query error:", err)
				continue
			}

			for taskRows.Next() {
				var t db.Task
				err := taskRows.Scan(&t.Name, &t.Description, &t.Status, &t.CreatedAt, &t.DoneAt)

				if err != nil {
					log.Println("Task scan error:", err)
					continue
				}
				s.Tasks = append(s.Tasks, t)
			}

			err = taskRows.Close()
			if err != nil {
				return
			}

			sessions = append(sessions, s)
		}

		err = tmpl.Execute(w, sessions)
		if err != nil {
			http.Error(w, "Render error", 500)
			log.Println("Template exec error:", err)
		}
	}
}

func makeAdminHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/admin/":
			tmpl, err := template.ParseFiles("static/admin.html")
			if err != nil {
				http.Error(w, "Template parsing error", 500)
				log.Println(err)
				return
			}

			rows, err := dbConn.Query("SELECT name, description, status FROM tasks WHERE status = 'todo'")
			if err != nil {
				log.Println("DB query error:", err)
				http.Error(w, "DB error", 500)
				return
			}
			defer func(rows *sql.Rows) {
				err := rows.Close()
				if err != nil {
					log.Println("Rows closing error:", err)
				}
			}(rows)

			var tasks []db.Task

			for rows.Next() {
				var t db.Task
				err := rows.Scan(&t.Name, &t.Description, &t.Status)
				if err != nil {
					log.Println("Row scan error:", err)
					continue
				}
				tasks = append(tasks, t)
			}

			if err = rows.Err(); err != nil {
				log.Println("Rows error:", err)
			}

			err = tmpl.Execute(w, struct {
				Tasks []db.Task
			}{
				Tasks: tasks,
			})

			if err != nil {
				http.Error(w, "Render error", 500)
				log.Println("Template exec error:", err)
			}
			return
		case "/admin/add-task":
			name := r.PostFormValue("name")
			description := r.PostFormValue("description")

			_, err := dbConn.Query(
				"INSERT INTO tasks(name, description, status) VALUES ($1, $2, 'todo')", name, description)
			if err != nil {
				log.Println("Error while creating a new tasks entry:", err)
				return
			}
			http.Redirect(w, r, "/admin/", http.StatusSeeOther)

			return

		case "/complete-task/":
			return
		}

	}
}

/*
func makeMainHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db.Exec("INSERT INTO temp DEFAULT VALUES")
	}
}
*/
