package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gazizov-ai/online-checkers/pkg/config"
	"github.com/gazizov-ai/online-checkers/pkg/db"
	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"
	appkafka "github.com/gazizov-ai/online-checkers/pkg/kafka"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/handler"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/identity"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/outbox"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/repository"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/service"
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

	userRepo := repository.NewPostgresUserRepository(database)

	if cfg.JWTPrivateKeyPath == "" {
		log.Fatal("JWT_PRIVATE_KEY_PATH is required")
	}

	privateKey, err := identity.LoadRSAPrivateKey(cfg.JWTPrivateKeyPath)
	if err != nil {
		log.Fatalf("load RSA private key: %v", err)
	}

	issuer := identity.NewRSAIssuer(
		privateKey,
		cfg.JWTKeyID,
		cfg.OIDCIssuer,
		cfg.OIDCAudience,
		cfg.AccessTokenTTL,
		cfg.IDTokenTTL,
	)

	authService := service.NewAuthService(userRepo, issuer, cfg.UserRegisteredTopic)
	authHandler := handler.NewAuthHandler(authService)

	identityHandler := handler.NewIdentityHandler(issuer)

	producer := appkafka.NewProducer(appkafka.ProducerConfig{
		Brokers: strings.Split(cfg.KafkaBrokers, ","),
		Topic:   cfg.UserRegisteredTopic,
	})
	defer producer.Close()

	outboxWorker := outbox.NewWorker(userRepo, producer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go outboxWorker.Run(ctx)

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

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		r.Group(func(r chi.Router) {
			r.Use(appjwt.Middleware(issuer))
			r.Get("/me", authHandler.Me)
		})
	})

	r.Get("/.well-known/jwks.json", identityHandler.JWKS)
	r.Get("/.well-known/openid-configuration", identityHandler.Discovery)

	addr := ":" + cfg.HTTPPort

	log.Printf("%s listening on %s", cfg.ServiceName, addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
