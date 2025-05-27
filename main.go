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
	secondHandler := makeSecondHandler()
	workLogHandler := makeWorkLogHandler(db)

	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/worklog/", workLogHandler)
	http.HandleFunc("/second/", secondHandler)

	log.Printf("Starting server at %s", addr)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Fatal(
		http.ListenAndServe(":8080", nil))
	log.Printf("db is %v", db)

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

		rows, err := dbConn.Query("SELECT id, date, start_time, end_time, duration FROM work_sessions ORDER BY date DESC")
		if err != nil {
			http.Error(w, "DB error", 500)
			log.Println(err)
			return
		}
		defer rows.Close()

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

			taskRows, err := dbConn.Query("SELECT name, description, status, created_at, done_at FROM tasks WHERE session_id = $1", id)

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
			taskRows.Close()

			sessions = append(sessions, s)
		}

		err = tmpl.Execute(w, sessions)
		if err != nil {
			http.Error(w, "Render error", 500)
			log.Println("Template exec error:", err)
		}
	}
}

func makeSecondHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		if r.URL.Path != "/second/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./static/second.html")
	}
}

/*
func makeMainHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db.Exec("INSERT INTO temp DEFAULT VALUES")
	}
}
*/
