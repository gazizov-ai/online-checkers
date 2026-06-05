package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	gamev1 "github.com/gazizov-ai/online-checkers/gen/game/v1"
	"github.com/gazizov-ai/online-checkers/pkg/config"
	"github.com/gazizov-ai/online-checkers/pkg/db"
	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"
	gamegrpc "github.com/gazizov-ai/online-checkers/services/game/internal/grpc"
	"github.com/gazizov-ai/online-checkers/services/game/internal/handler"
	"github.com/gazizov-ai/online-checkers/services/game/internal/repository"
	"github.com/gazizov-ai/online-checkers/services/game/internal/service"
	gamews "github.com/gazizov-ai/online-checkers/services/game/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
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

	gameRepo := repository.NewPostgresGameRepository(database)
	gameService := service.NewGameService(gameRepo)
	roomManager := gamews.NewRoomManager()

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

	gameHandler := handler.NewHandler(gameService, roomManager, tokenVerifier)

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

	gameHandler.RegisterRoutes(r)

	grpcServer := grpc.NewServer()

	gamev1.RegisterGameServiceServer(
		grpcServer,
		gamegrpc.NewGameServer(gameService),
	)

	grpcAddr := ":" + cfg.GRPCPort

	grpcListener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("listen grpc: %v", err)
	}

	go func() {
		log.Printf("%s gRPC listening on %s", cfg.ServiceName, grpcAddr)

		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("grpc server error: %v", err)
		}
	}()

	addr := ":" + cfg.HTTPPort

	log.Printf("%s listening on %s", cfg.ServiceName, addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server error: %v", err)
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
