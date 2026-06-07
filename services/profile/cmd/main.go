package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gazizov-ai/online-checkers/pkg/config"
	"github.com/gazizov-ai/online-checkers/pkg/db"
	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"
	appkafka "github.com/gazizov-ai/online-checkers/pkg/kafka"
	profileconsumer "github.com/gazizov-ai/online-checkers/services/profile/internal/consumer"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/handler"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/repository"
	"github.com/gazizov-ai/online-checkers/services/profile/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/lib/pq"
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

	profileRepo := repository.NewPostgresProfileRepository(database)
	profileService := service.NewProfileService(profileRepo)
	profileHandler := handler.NewProfileHandler(profileService)

	userRegisteredHandler := profileconsumer.NewUserRegisteredHandler(profileService)

	kafkaConsumer := appkafka.NewConsumer(appkafka.ConsumerConfig{
		Brokers: strings.Split(cfg.KafkaBrokers, ","),
		Topic:   cfg.UserRegisteredTopic,
		GroupID: cfg.ProfileConsumerGroup,
	})
	defer kafkaConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			if err := kafkaConsumer.Run(ctx, func(ctx context.Context, msg appkafka.Message) error {
				return userRegisteredHandler.Handle(ctx, msg.Value, msg.Headers)
			}); err != nil {
				if ctx.Err() != nil {
					return
				}

				log.Printf("profile user.registered consumer stopped: %v", err)
				time.Sleep(2 * time.Second)
			}
		}
	}()

	jwksCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	publicKey, err := loadJWKSWithRetry(
		jwksCtx,
		cfg.JWKSURL,
		cfg.JWTKeyID,
		10,
		2*time.Second,
	)
	if err != nil {
		log.Fatalf("load jwks: %v", err)
	}

	tokenVerifier := appjwt.NewRS256Verifier(
		publicKey,
		cfg.JWTKeyID,
		cfg.OIDCIssuer,
		cfg.OIDCAudience,
	)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
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
			_ = httpx.WriteError(w, http.StatusServiceUnavailable, "not_ready", "database can't be reached")
			return
		}

		_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "ready",
			"service": cfg.ServiceName,
		})
	})

	r.Route("/api/v1/profiles", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(appjwt.Middleware(tokenVerifier))

			r.Get("/me", profileHandler.GetMyProfile)
			r.Patch("/me", profileHandler.UpdateMyProfile)
		})

		r.Post("/batch", profileHandler.BatchProfiles)
		r.Get("/{user_id}", profileHandler.GetProfileByUserID)
	})

	addr := ":" + cfg.HTTPPort
	log.Printf("%s listening on %s", cfg.ServiceName, addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("http server: %v", err)
	}
}

func loadJWKSWithRetry(
	ctx context.Context,
	jwksURL string,
	keyID string,
	attempts int,
	delay time.Duration,
) (*rsa.PublicKey, error) {
	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		publicKey, err := appjwt.LoadRSAPublicKeyFromJWKS(ctx, jwksURL, keyID)
		if err == nil {
			return publicKey, nil
		}

		lastErr = err

		if attempt == attempts {
			break
		}

		log.Printf("load jwks failed, retrying: attempt=%d/%d error=%v", attempt, attempts, err)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	return nil, fmt.Errorf("load jwks after %d attempts: %w", attempts, lastErr)
}
