package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gazizov-ai/online-checkers/pkg/config"
	"github.com/gazizov-ai/online-checkers/pkg/db"
	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer database.Close()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	//r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": cfg.ServiceName,
		})
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := database.PingContext(ctx); err != nil {
			_ = httpx.WriteError(
				w,
				http.StatusServiceUnavailable,
				"not_ready",
				"database can't be reached",
			)
			return
		}

		_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "ready",
			"service": cfg.ServiceName,
		})
	})

	addr := ":" + cfg.HTTPPort

	log.Printf("%s listening on %s", cfg.ServiceName, addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
