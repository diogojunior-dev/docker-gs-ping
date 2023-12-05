package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, Docker! <3\n")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct{ Status string }{Status: "OK"})
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
