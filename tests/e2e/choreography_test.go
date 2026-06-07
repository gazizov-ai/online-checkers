//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	eventsv1 "github.com/gazizov-ai/online-checkers/gen/events/v1"
	appkafka "github.com/gazizov-ai/online-checkers/pkg/kafka"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestUserRegisteredCreatesProfile(t *testing.T) {
	authURL := envOrDefault("E2E_AUTH_URL", "http://localhost:8081")
	profileURL := envOrDefault("E2E_PROFILE_URL", "http://localhost:8086")
	username := "e2e_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:12]

	var registered struct {
		User struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"user"`
	}
	postJSON(t, authURL+"/api/v1/register", map[string]string{
		"username": username,
		"password": "e2e-password",
	}, http.StatusCreated, &registered)

	var profile struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
	}
	eventually(t, 20*time.Second, func() (bool, error) {
		status, err := getJSON(profileURL+"/api/v1/profiles/"+registered.User.ID, &profile)
		if err != nil {
			return false, err
		}
		return status == http.StatusOK, nil
	})

	if profile.UserID != registered.User.ID || profile.Username != username {
		t.Fatalf("profile = %+v, registered user = %+v", profile, registered.User)
	}
}

func TestGameFinishedUpdatesRatings(t *testing.T) {
	ratingURL := envOrDefault("E2E_RATING_URL", "http://localhost:8084")
	brokers := strings.Split(envOrDefault("E2E_KAFKA_BROKERS", "localhost:9092"), ",")
	topic := envOrDefault("E2E_GAME_FINISHED_TOPIC", "game.finished")
	white := uuid.New()
	black := uuid.New()
	eventID := uuid.New()

	message := &eventsv1.GameFinished{
		EventId:       eventID.String(),
		GameId:        uuid.NewString(),
		WhitePlayerId: white.String(),
		BlackPlayerId: black.String(),
		WinnerId:      white.String(),
		Result:        "white_win",
		Reason:        "checkers_rules",
		FinishedAt:    timestamppb.Now(),
	}
	payload, err := proto.Marshal(message)
	if err != nil {
		t.Fatalf("marshal game.finished: %v", err)
	}

	producer := appkafka.NewProducer(appkafka.ProducerConfig{Brokers: brokers, Topic: topic})
	t.Cleanup(func() { _ = producer.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := producer.Publish(ctx, []byte(message.GameId), payload, map[string]string{
		"event_type":   "game.finished",
		"content_type": "application/protobuf",
	}); err != nil {
		t.Fatalf("publish game.finished: %v", err)
	}

	var winnerRating ratingResponse
	eventually(t, 20*time.Second, func() (bool, error) {
		status, err := getJSON(ratingURL+"/api/v1/ratings/"+white.String(), &winnerRating)
		if err != nil {
			return false, err
		}
		return status == http.StatusOK, nil
	})

	var loserRating ratingResponse
	status, err := getJSON(ratingURL+"/api/v1/ratings/"+black.String(), &loserRating)
	if err != nil || status != http.StatusOK {
		t.Fatalf("get loser rating: status=%d error=%v", status, err)
	}
	if winnerRating.Rating != 1025 || winnerRating.Wins != 1 || winnerRating.GamesPlayed != 1 {
		t.Fatalf("winner rating = %+v", winnerRating)
	}
	if loserRating.Rating != 975 || loserRating.Losses != 1 || loserRating.GamesPlayed != 1 {
		t.Fatalf("loser rating = %+v", loserRating)
	}
}

type ratingResponse struct {
	UserID      string `json:"user_id"`
	Rating      int    `json:"rating"`
	GamesPlayed int    `json:"games_played"`
	Wins        int    `json:"wins"`
	Losses      int    `json:"losses"`
}

func postJSON(t *testing.T, target string, body any, wantStatus int, response any) {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	resp, err := http.Post(target, "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST %s: %v", target, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		contents, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST %s status=%d, want=%d body=%s", target, resp.StatusCode, wantStatus, contents)
	}
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		t.Fatalf("decode POST %s response: %v", target, err)
	}
}

func getJSON(target string, response any) (int, error) {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(target)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return resp.StatusCode, err
		}
	}
	return resp.StatusCode, nil
}

func eventually(t *testing.T, timeout time.Duration, check func() (bool, error)) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		ok, err := check()
		if ok {
			return
		}
		if err != nil {
			lastErr = err
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("condition not met within %s: %v", timeout, lastErr)
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func TestMain(m *testing.M) {
	for name, target := range map[string]string{
		"auth":    envOrDefault("E2E_AUTH_URL", "http://localhost:8081") + "/health",
		"profile": envOrDefault("E2E_PROFILE_URL", "http://localhost:8086") + "/health",
		"rating":  envOrDefault("E2E_RATING_URL", "http://localhost:8084") + "/health",
	} {
		resp, err := (&http.Client{Timeout: 2 * time.Second}).Get(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "e2e prerequisite %s is unavailable at %s: %v\n", name, target, err)
			os.Exit(1)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "e2e prerequisite %s returned HTTP %d\n", name, resp.StatusCode)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}
