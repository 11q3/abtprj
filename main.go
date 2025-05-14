package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

const addr = "http://localhost:8080"
const connStr = "postgres://postgres:1121231@localhost:5432/abtprj?sslmode=disable"

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello world!")
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Println("error while connecting to db")
		return
	}
	res, err := db.Exec("INSERT INTO temp DEFAULT VALUES")
	log.Println(res)
	log.Println(err)

}

func main() {
	fmt.Println("Hello, World!")

	http.HandleFunc("/", handler)
	log.Printf("Starting server at %s", addr)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
