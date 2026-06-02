package main

import (
	"fmt"
	"log"
	"net/http"

	"linkup/config"
	"linkup/db"
)

func main() {
	if err := config.LoadEnv(); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	env := config.GetEnv()
	port := env.Port

	database, err := db.ConnectDb(env)
	if err != nil {
		log.Printf("DB connection: failed (%v)", err)
	} else {
		log.Println("DB connection: success")
		defer database.Close()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("LinkUp server is running"))
	})

	addr := ":" + port
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
