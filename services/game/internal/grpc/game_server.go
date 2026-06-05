package grpc

import (
	"context"

	gamev1 "github.com/gazizov-ai/online-checkers/gen/game/v1"
	"github.com/gazizov-ai/online-checkers/services/game/internal/service"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GameServer struct {
	gamev1.UnimplementedGameServiceServer

	gameService *service.GameService
}

func NewGameServer(gameService *service.GameService) *GameServer {
	return &GameServer{
		gameService: gameService,
	}
}

func (s *GameServer) CreateGame(
	ctx context.Context,
	req *gamev1.CreateGameRequest,
) (*gamev1.CreateGameResponse, error) {
	whiteID, err := uuid.Parse(req.GetWhitePlayerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid white player id")
	}

	blackID, err := uuid.Parse(req.GetBlackPlayerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid black player id")
	}

	if whiteID == blackID {
		return nil, status.Error(codes.InvalidArgument, "players must be different")
	}

	gameID, err := s.gameService.CreateGame(ctx, service.CreateGameInput{
		WhitePlayerID: whiteID,
		BlackPlayerID: blackID,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create game")
	}

	return &gamev1.CreateGameResponse{
		GameId: gameID.String(),
	}, nil
}
