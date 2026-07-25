package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"socratic-tutor-harness/internal/tutor"
	"time"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", "0.0.0.0:8083", "地址")
	flag.Parse()

	dataDir := "data"
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		fmt.Printf("The dir didn't make,detail: %s", err)
		return
	}

	dbPath := filepath.Join(dataDir, "tutor.db")
	db, err := tutor.BuildDatabase(dbPath)
	if err != nil {
		fmt.Printf("The database didn't build,detail: %s", err)
		return
	}

	defer db.Close()

	socraticServer := &tutor.SocraticServer{DB: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/socratic/ask", socraticServer.HandleAsk)
	mux.HandleFunc("/healthz", socraticServer.HandleHealthz)

	httpServer := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 130 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("The tutor server on %s", httpServer.Addr)
	err = httpServer.ListenAndServe()
	if err != nil {
		fmt.Printf("The server didn't listen on %s,detail: %s", httpServer.Addr, err)
	}
}
