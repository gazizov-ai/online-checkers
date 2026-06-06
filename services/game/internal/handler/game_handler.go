package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gazizov-ai/online-checkers/pkg/httpx"
	appjwt "github.com/gazizov-ai/online-checkers/pkg/jwt"
	"github.com/gazizov-ai/online-checkers/services/game/internal/checkers"
	"github.com/gazizov-ai/online-checkers/services/game/internal/domain"
	"github.com/gazizov-ai/online-checkers/services/game/internal/service"
	gamews "github.com/gazizov-ai/online-checkers/services/game/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	gorilla "github.com/gorilla/websocket"
)

type createGameRequest struct {
	WhitePlayerID string `json:"white_player_id"`
	BlackPlayerID string `json:"black_player_id"`
}

type createGameResponse struct {
	GameID string `json:"game_id"`
}

type getGameResponse struct {
	ID            string                `json:"id"`
	WhitePlayerID string                `json:"white_player_id"`
	BlackPlayerID string                `json:"black_player_id"`
	Status        string                `json:"status"`
	WinnerID      *string               `json:"winner_id,omitempty"`
	BoardState    checkers.GameSnapshot `json:"board_state"`
	CurrentTurn   string                `json:"current_turn"`
	Result        *string               `json:"result,omitempty"`
	FinishReason  *string               `json:"finish_reason,omitempty"`
	DrawOfferBy   *string               `json:"draw_offer_by,omitempty"`
}

type Handler struct {
	gameService   *service.GameService
	roomManager   *gamews.RoomManager
	tokenVerifier appjwt.Verifier
}

func NewHandler(
	gameService *service.GameService,
	roomManager *gamews.RoomManager,
	tokenVerifier appjwt.Verifier,
) *Handler {
	return &Handler{
		gameService:   gameService,
		roomManager:   roomManager,
		tokenVerifier: tokenVerifier,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/api/v1/games", h.createGame)
	r.Get("/api/v1/games/{game_id}", h.getGame)
	r.Get("/api/v1/games/{game_id}/ws", h.connectWebSocket)
}

var upgrader = gorilla.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) createGame(w http.ResponseWriter, r *http.Request) {
	var req createGameRequest

	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid request body")
		return
	}

	whiteID, err := uuid.Parse(req.WhitePlayerID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_white_id", "invalid white player id")
		return
	}

	blackID, err := uuid.Parse(req.BlackPlayerID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_black_id", "invalid black player id")
		return
	}

	if whiteID == blackID {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "players must be different")
		return
	}

	gameID, err := h.gameService.CreateGame(r.Context(), service.CreateGameInput{
		WhitePlayerID: whiteID,
		BlackPlayerID: blackID,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "game_create_failed", "failed to create game")
		return
	}

	resp := createGameResponse{
		GameID: gameID.String(),
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) getGame(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(chi.URLParam(r, "game_id"))
	if err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_game_id", "invalid game id")
		return
	}

	game, err := h.gameService.GetGame(r.Context(), gameID)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusNotFound, "game_not_found", "game not found")
		return
	}

	resp := getGameResponse{
		ID:            gameID.String(),
		WhitePlayerID: game.WhitePlayerID.String(),
		BlackPlayerID: game.BlackPlayerID.String(),
		Status:        string(game.Status),
		WinnerID:      uuidString(game.WinnerID),
		BoardState:    game.Snapshot,
		CurrentTurn:   string(game.CurrentTurn),
		Result:        gameResultString(game.Result),
		FinishReason:  finishReasonString(game.FinishReason),
		DrawOfferBy:   uuidString(game.DrawOfferBy),
	}

	_ = httpx.WriteJSON(w, http.StatusOK, resp)
}

func winnerIDString(gameWinnerID *uuid.UUID) *string {
	if gameWinnerID == nil {
		return nil
	}

	s := gameWinnerID.String()
	return &s
}

func uuidString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}

	s := value.String()
	return &s
}

func gameResultString(value *domain.GameResult) *string {
	if value == nil {
		return nil
	}

	s := string(*value)
	return &s
}

func finishReasonString(value *domain.FinishReason) *string {
	if value == nil {
		return nil
	}

	s := string(*value)
	return &s
}

func gameStatePayload(game domain.Game) gamews.GameStatePayload {
	return gamews.GameStatePayload{
		GameID:       game.ID.String(),
		BoardState:   game.Snapshot,
		Status:       string(game.Status),
		CurrentTurn:  string(game.CurrentTurn),
		WinnerID:     uuidString(game.WinnerID),
		Result:       gameResultString(game.Result),
		FinishReason: finishReasonString(game.FinishReason),
		DrawOfferBy:  uuidString(game.DrawOfferBy),
	}
}

func (h *Handler) connectWebSocket(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(chi.URLParam(r, "game_id"))
	if err != nil {
		_ = httpx.WriteError(w, http.StatusBadRequest, "invalid_game_id", "invalid game id")
		return
	}

	rawToken, ok := appjwt.TokenFromRequest(r)
	if !ok {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}

	claims, err := h.tokenVerifier.VerifyAccessToken(r.Context(), rawToken)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusUnauthorized, "invalid_token", "invalid bearer token")
		return
	}

	userID := claims.UserID

	game, err := h.gameService.GetGame(r.Context(), gameID)
	if err != nil {
		_ = httpx.WriteError(w, http.StatusNotFound, "game_not_found", "game not found")
		return
	}

	if userID != game.WhitePlayerID && userID != game.BlackPlayerID {
		_ = httpx.WriteError(w, http.StatusForbidden, "forbidden", "user is not a player of this game")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	userIDString := userID.String()
	client := gamews.NewClient(userIDString, conn)
	room := h.roomManager.GetOrCreateRoom(gameID.String())
	room.AddClient(client)

	err = client.Send(gamews.NewGameStateMessage(gameStatePayload(game)))
	if err != nil {
		room.RemoveClient(userIDString)
		_ = client.Close()
		return
	}

	defer func() {
		room.RemoveClient(userIDString)
		_ = client.Close()
	}()

	for {
		var msg gamews.IncomingMessage

		if err := conn.ReadJSON(&msg); err != nil {
			return
		}

		switch msg.Type {
		case gamews.MessageTypeMove:
			if err := h.handleWSMove(r.Context(), room, gameID, userID, msg.Payload); err != nil {
				_ = client.Send(gamews.NewErrorMessage(err.Error()))
				continue
			}
		case gamews.MessageTypeResign:
			if err := h.handleWSResign(r.Context(), room, gameID, userID); err != nil {
				_ = client.Send(gamews.NewErrorMessage(err.Error()))
				continue
			}

		case gamews.MessageTypeDrawOffer:
			if err := h.handleWSDrawOffer(r.Context(), room, gameID, userID); err != nil {
				_ = client.Send(gamews.NewErrorMessage(err.Error()))
				continue
			}

		case gamews.MessageTypeDrawResponse:
			if err := h.handleWSDrawResponse(r.Context(), room, gameID, userID, msg.Payload); err != nil {
				_ = client.Send(gamews.NewErrorMessage(err.Error()))
				continue
			}
		default:
			_ = client.Send(gamews.NewErrorMessage("unknown message type"))
		}
	}
}

func (h *Handler) handleWSMove(
	ctx context.Context,
	room *gamews.GameRoom,
	gameID uuid.UUID,
	playerID uuid.UUID,
	payload json.RawMessage,
) error {
	var movePayload gamews.MovePayload
	if err := json.Unmarshal(payload, &movePayload); err != nil {
		return err
	}

	output, err := h.gameService.ApplyMove(ctx, service.ApplyMoveInput{
		GameID:   gameID,
		PlayerID: playerID,
		From:     movePayload.From,
		To:       movePayload.To,
	})
	if err != nil {
		return err
	}

	game := output.Game

	room.Broadcast(gamews.NewGameStateMessage(gameStatePayload(game)))

	if game.Status == domain.GameStatusFinished {
		room.Broadcast(gamews.NewGameFinishedMessage(gamews.GameFinishedPayload{
			GameID:       gameID.String(),
			WinnerID:     uuidString(game.WinnerID),
			Result:       gameResultString(game.Result),
			FinishReason: finishReasonString(game.FinishReason),
		}))

		room.CloseAll()
		h.roomManager.DeleteRoom(gameID.String())
	}

	return nil
}

func (h *Handler) handleWSResign(
	ctx context.Context,
	room *gamews.GameRoom,
	gameID uuid.UUID,
	playerID uuid.UUID,
) error {
	output, err := h.gameService.Resign(ctx, service.ResignInput{
		GameID:   gameID,
		PlayerID: playerID,
	})
	if err != nil {
		return err
	}

	game := output.Game

	room.Broadcast(gamews.NewGameStateMessage(gameStatePayload(game)))

	room.Broadcast(gamews.NewGameFinishedMessage(gamews.GameFinishedPayload{
		GameID:       gameID.String(),
		WinnerID:     uuidString(game.WinnerID),
		Result:       gameResultString(game.Result),
		FinishReason: finishReasonString(game.FinishReason),
	}))

	room.CloseAll()
	h.roomManager.DeleteRoom(gameID.String())

	return nil
}

func (h *Handler) handleWSDrawOffer(
	ctx context.Context,
	room *gamews.GameRoom,
	gameID uuid.UUID,
	playerID uuid.UUID,
) error {
	output, err := h.gameService.OfferDraw(ctx, service.OfferDrawInput{
		GameID:   gameID,
		PlayerID: playerID,
	})
	if err != nil {
		return err
	}

	game := output.Game

	room.Broadcast(gamews.NewGameStateMessage(gameStatePayload(game)))

	room.Broadcast(gamews.NewDrawOfferedMessage(gamews.DrawOfferedPayload{
		GameID:    game.ID.String(),
		OfferedBy: playerID.String(),
	}))

	return nil
}

func (h *Handler) handleWSDrawResponse(
	ctx context.Context,
	room *gamews.GameRoom,
	gameID uuid.UUID,
	playerID uuid.UUID,
	payload json.RawMessage,
) error {
	var req gamews.DrawResponsePayload
	if err := json.Unmarshal(payload, &req); err != nil {
		return err
	}

	output, err := h.gameService.RespondDraw(ctx, service.RespondDrawInput{
		GameID:   gameID,
		PlayerID: playerID,
		Accepted: req.Accepted,
	})
	if err != nil {
		return err
	}

	game := output.Game

	room.Broadcast(gamews.NewGameStateMessage(gameStatePayload(game)))

	if !req.Accepted {
		room.Broadcast(gamews.NewDrawDeclinedMessage(gamews.DrawDeclinedPayload{
			GameID:     game.ID.String(),
			DeclinedBy: playerID.String(),
		}))
		return nil
	}

	room.Broadcast(gamews.NewGameFinishedMessage(gamews.GameFinishedPayload{
		GameID:       game.ID.String(),
		WinnerID:     uuidString(game.WinnerID),
		Result:       gameResultString(game.Result),
		FinishReason: finishReasonString(game.FinishReason),
	}))

	room.CloseAll()
	h.roomManager.DeleteRoom(gameID.String())

	return nil
}
