package main

import (
	"database/sql"
	"log"
	"net/http"

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

	handler := makeHandler()

	http.HandleFunc("/", handler)

	log.Printf("Starting server at %s", addr)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Fatal(
		http.ListenAndServe(":8080", nil))
	log.Printf("db is %v", db)

}

func makeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "./static/index.html")
	}
}

/*
func makeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db.Exec("INSERT INTO temp DEFAULT VALUES")
	}
}
*/
