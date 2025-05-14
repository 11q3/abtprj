package main

import (
	"fmt"
	"log"
	"net/http"
)

const addr = "http://localhost:8080"

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "hello world!")
}

func main() {
	fmt.Println("Hello, World!")

	http.HandleFunc("/", handler) //prefer it to be /
	log.Printf("Starting server at %s", addr)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
