package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gazizov-ai/online-checkers/services/auth/internal/domain"
	"github.com/gazizov-ai/online-checkers/services/auth/internal/repository"
	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

type AuthService struct {
	users  UserRepository
	tokens TokenIssuer
}

func NewAuthService(users UserRepository, tokens TokenIssuer) *AuthService {
	return &AuthService{
		users:  users,
		tokens: tokens,
	}
}

type RegisterInput struct {
	Username string
	Email    *string
	Password string
}

type LoginInput struct {
	Username string
	Password string
}

type LoginResult struct {
	User   *domain.User
	Tokens *TokenPair
}

type TokenSubject struct {
	UserID   uuid.UUID
	Username string
	Email    *string
}

type TokenPair struct {
	AccessToken string
	IDToken     string
	TokenType   string
	ExpiresIn   int64
}

type TokenIssuer interface {
	IssueTokens(ctx context.Context, subject TokenSubject) (*TokenPair, error)
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*domain.User, error) {
	if input.Username == "" {
		return nil, ErrInvalidUsername
	}

	if len(input.Password) < 8 {
		return nil, ErrInvalidPassword
	}

	passwordHash, err := hashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := domain.User{
		ID:           uuid.New(),
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.users.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUsernameAlreadyExists) {
			return nil, ErrUsernameAlreadyTaken
		}

		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			return nil, ErrEmailAlreadyTaken
		}

		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	if input.Username == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.GetUserByUsername(ctx, input.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf("get user by username: %w", err)
	}

	if err := checkPassword(user.PasswordHash, input.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokens, err := s.tokens.IssueTokens(ctx, TokenSubject{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("issue tokens: %w", err)
	}

	return &LoginResult{
		User:   user,
		Tokens: tokens,
	}, nil
}

func (s *AuthService) Me(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}
