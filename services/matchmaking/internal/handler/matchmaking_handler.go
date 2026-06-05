package handler

import (
	"errors"
	"net/http"

	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"
	"github.com/gazizov-ai/online-checkers/services/matchmaking/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	matchmakingService *service.MatchmakingService
}

func NewHandler(matchmakingService *service.MatchmakingService) *Handler {
	return &Handler{
		matchmakingService: matchmakingService,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router, jwtMiddleware func(http.Handler) http.Handler) {
	r.Route("/api/v1/matchmaking", func(r chi.Router) {
		r.Use(jwtMiddleware)

		r.Post("/search", h.Search)
		r.Get("/status", h.Status)
		r.Post("/cancel", h.Cancel)
	})
}

type MatchmakingResponse struct {
	Status string  `json:"status"`
	GameID *string `json:"game_id,omitempty"`
}

func toMatchmakingResponse(result service.SearchResult) MatchmakingResponse {
	var gameID *string
	if result.GameID != nil {
		value := result.GameID.String()
		gameID = &value
	}

	return MatchmakingResponse{
		Status: string(result.Status),
		GameID: gameID,
	}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	userID, ok := appjwt.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	result, err := h.matchmakingService.Search(r.Context(), userID)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to search match")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, toMatchmakingResponse(result))
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	userID, ok := appjwt.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	result, err := h.matchmakingService.Status(r.Context(), userID)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to get matchmaking status")
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, toMatchmakingResponse(result))
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID, ok := appjwt.UserIDFromContext(r.Context())
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	if err := h.matchmakingService.Cancel(r.Context(), userID); err != nil {
		if errors.Is(err, service.ErrAlreadyMatched) {
			_ = httpx.WriteError(
				w,
				http.StatusConflict,
				"already_matched",
				"match has already been created",
			)
			return
		}

		_ = httpx.WriteError(
			w,
			http.StatusInternalServerError,
			"internal_error",
			"failed to cancel matchmaking",
		)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "cancelled",
	})
}
