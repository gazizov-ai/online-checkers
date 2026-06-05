package client

import (
	"context"

	gamev1 "github.com/gazizov-ai/online-checkers/gen/game/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GameClient struct {
	conn *grpc.ClientConn

	game gamev1.GameServiceClient
}

func NewGameClient(addr string) (*GameClient, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &GameClient{
		conn: conn,
		game: gamev1.NewGameServiceClient(conn),
	}, nil
}

func (c *GameClient) Close() error {
	return c.conn.Close()
}

func (c *GameClient) CreateGame(
	ctx context.Context,
	whiteID uuid.UUID,
	blackID uuid.UUID,
) (uuid.UUID, error) {
	resp, err := c.game.CreateGame(ctx, &gamev1.CreateGameRequest{
		WhitePlayerId: whiteID.String(),
		BlackPlayerId: blackID.String(),
	})
	if err != nil {
		return uuid.Nil, err
	}

	gameID, err := uuid.Parse(resp.GetGameId())
	if err != nil {
		return uuid.Nil, err
	}

	return gameID, nil
}
