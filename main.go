package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cenkalti/backoff/v4"
	_ "github.com/cockroachdb/cockroach-go/v2/crdb"
	_ "github.com/lib/pq"
)

func main() {
	db, err := initStore()
	if err != nil {
		log.Fatal("Failed to initialise the store: %s", err)
		return
	}

	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			resp["error"] = "Method not allowed"
			json.NewEncoder(w).Encode(resp)
			return
		}

		v, err := rootHandler(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp["error"] = err.Error()
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusOK)
		resp["message"] = v
		json.NewEncoder(w).Encode(resp)
		return
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			resp["error"] = "Method not allowed"
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusOK)
		resp["status"] = "OK"
		json.NewEncoder(w).Encode(resp)
		return
	})

	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			resp["error"] = "Method not allowed"
			json.NewEncoder(w).Encode(resp)
			return
		}

		var msg Message
		json.NewDecoder(r.Body).Decode(&msg)

		if err := sendHandler(db, &msg); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp["error"] = err.Error()
			json.NewEncoder(w).Encode(resp)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(msg)
		return
	})

	log.Println("Tudo pronto!")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Message struct {
	Value string `json:"value"`
}

func initStore() (*sql.DB, error) {
	pgConnString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
	)

	var (
		db  *sql.DB
		err error
	)

	openDB := func() error {
		db, err = sql.Open("postgres", pgConnString)
		return err
	}

	err = backoff.Retry(openDB, backoff.NewExponentialBackOff())

	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS message (value TEXT PRIMARY KEY)")

	if err != nil {
		return nil, err
	}

	return db, nil
}

func rootHandler(db *sql.DB) (string, error) {
	recCount, err := countRecords(db)

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Hello, Docker! (%d)\n", recCount), nil
}

func sendHandler(db *sql.DB, msg *Message) error {
	_, err := db.Exec("INSERT INTO message (value) VALUES ($1) ON CONFLICT (value) DO UPDATE SET value = excluded.value", msg.Value)
	return err
}

func countRecords(db *sql.DB) (int, error) {
	rows, err := db.Query("SELECT COUNT(*) FROM message")

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	count := 0

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
		rows.Close()
	}

	return count, nil
}
