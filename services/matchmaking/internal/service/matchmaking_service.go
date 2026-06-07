package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gazizov-ai/online-checkers/services/matchmaking/internal/domain"
	"github.com/gazizov-ai/online-checkers/services/matchmaking/internal/repository"
	"github.com/google/uuid"
)

type GameCreator interface {
	CreateGame(ctx context.Context, whiteID uuid.UUID, blackID uuid.UUID) (uuid.UUID, error)
}

type MatchmakingService struct {
	repo       repository.QueueRepository
	gameClient GameCreator
}

func NewMatchmakingService(
	repo repository.QueueRepository,
	gameClient GameCreator,
) *MatchmakingService {
	return &MatchmakingService{
		repo:       repo,
		gameClient: gameClient,
	}
}

type SearchResult struct {
	Status domain.SearchStatus
	GameID *uuid.UUID
}

func (s *MatchmakingService) Search(
	ctx context.Context,
	userID uuid.UUID,
) (SearchResult, error) {
	_ = s.repo.CleanupStaleReservations(ctx, time.Minute)

	reservation, err := s.repo.EnqueueOrReservePair(ctx, userID)
	if err != nil {
		return SearchResult{}, err
	}

	switch reservation.Status {
	case domain.StatusWaiting:
		return SearchResult{
			Status: domain.StatusWaiting,
			GameID: nil,
		}, nil

	case domain.StatusMatched:
		return SearchResult{
			Status: domain.StatusMatched,
			GameID: reservation.GameID,
		}, nil

	case domain.StatusMatching:
		if reservation.OpponentID == nil {
			return SearchResult{
				Status: domain.StatusWaiting,
				GameID: nil,
			}, nil
		}

		gameID, err := s.gameClient.CreateGame(ctx, *reservation.OpponentID, userID)
		if err != nil {
			_ = s.repo.ReleaseReservation(ctx, userID, *reservation.OpponentID)
			return SearchResult{}, err
		}

		if err := s.repo.CompleteMatch(ctx, userID, *reservation.OpponentID, gameID); err != nil {
			_ = s.repo.ReleaseReservation(ctx, userID, *reservation.OpponentID)
			return SearchResult{}, err
		}

		return SearchResult{
			Status: domain.StatusMatched,
			GameID: &gameID,
		}, nil

	default:
		return SearchResult{}, fmt.Errorf("unknown matchmaking status: %s", reservation.Status)
	}
}

func (s *MatchmakingService) Status(
	ctx context.Context,
	userID uuid.UUID,
) (SearchResult, error) {
	matchedEntry, err := s.repo.ConsumeMatchedByUserID(ctx, userID)
	if err != nil {
		return SearchResult{}, err
	}

	if matchedEntry != nil {
		return SearchResult{
			Status: matchedEntry.Status,
			GameID: matchedEntry.GameID,
		}, nil
	}

	entry, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return SearchResult{}, err
	}

	if entry == nil {
		return SearchResult{
			Status: domain.StatusWaiting,
			GameID: nil,
		}, nil
	}

	if entry.Status == domain.StatusMatching {
		return SearchResult{
			Status: domain.StatusWaiting,
			GameID: nil,
		}, nil
	}

	return SearchResult{
		Status: entry.Status,
		GameID: entry.GameID,
	}, nil
}

func (s *MatchmakingService) Cancel(
	ctx context.Context,
	userID uuid.UUID,
) error {
	entry, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if entry == nil {
		return nil
	}

	if entry.Status == domain.StatusMatching {
		return ErrAlreadyMatched
	}

	return s.repo.Cancel(ctx, userID)
}
