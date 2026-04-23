package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"plan-write-service/internal/api"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "HTTP listen address")
	dataRoot := flag.String("data-root", defaultDataRoot(), "Path to the Hugo project root")
	flag.Parse()

	logger := log.New(os.Stdout, "plan-write-service ", log.LstdFlags|log.Lmsgprefix)
	handler := api.NewServer(api.Config{
		DataRoot: *dataRoot,
		Logger:   logger,
	})

	server := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Printf("listening on %s (data root: %s)", *addr, *dataRoot)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
}

func defaultDataRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	if filepath.Base(wd) == "plan-write-service" {
		return filepath.Clean(filepath.Join(wd, "..", ".."))
	}

	return wd
}
