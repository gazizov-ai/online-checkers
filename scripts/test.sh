#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODE="${1:-unit}"
COMPOSE=(docker compose -p online-checkers-test -f "$ROOT_DIR/deploy/docker-compose.test.yml")
DATABASE_URL="${TEST_DATABASE_URL:-postgres://postgres:postgres@localhost:55432/checkers?sslmode=disable}"
GOCACHE="${GOCACHE:-/tmp/online-checkers-go-build}"
POSTMAN_ENV_FILE="${POSTMAN_ENV_FILE:-/tmp/online-checkers-postman-environment.json}"
POSTMAN_RESULTS_DIR="${POSTMAN_RESULTS_DIR:-$ROOT_DIR/test-results/newman}"
MANAGED_STACK=0
GENERATED_JWT_KEY=0

export GOCACHE
export TEST_DATABASE_URL="$DATABASE_URL"

usage() {
  cat <<'EOF'
Usage: scripts/test.sh [unit|integration|e2e|postman|all]

Environment:
  TEST_DATABASE_URL       PostgreSQL used by integration tests.
  TEST_MANAGE_STACK=0     Do not start compose services for e2e tests.
  TEST_KEEP_STACK=1       Keep test compose services after the run.
  TEST_JWT_PRIVATE_KEY_PATH
                          RSA private key mounted into the auth service.
  POSTMAN_RESULTS_DIR     Newman JUnit output directory.
  E2E_AUTH_URL            Auth service base URL (default http://localhost:8081).
  E2E_MATCHMAKING_URL     Matchmaking service URL (default http://localhost:8082).
  E2E_GAME_URL            Game service base URL (default http://localhost:8083).
  E2E_PROFILE_URL         Profile service base URL (default http://localhost:8086).
  E2E_RATING_URL          Rating service base URL (default http://localhost:8084).
  E2E_KAFKA_BROKERS       Kafka brokers (default localhost:9092).
EOF
}

cleanup() {
  local exit_code=$?
  if [[ "$MANAGED_STACK" == "1" && "${TEST_KEEP_STACK:-0}" != "1" ]]; then
    "${COMPOSE[@]}" down -v >/dev/null 2>&1 || true
  fi
  if [[ "$GENERATED_JWT_KEY" == "1" ]]; then
    rm -f "$TEST_JWT_PRIVATE_KEY_PATH"
  fi
  rm -f "$POSTMAN_ENV_FILE"
  exit "$exit_code"
}

trap cleanup EXIT INT TERM

wait_for_postgres() {
  local attempts=60
  until psql "$DATABASE_URL" -c "SELECT 1" >/dev/null 2>&1; do
    attempts=$((attempts - 1))
    if [[ "$attempts" -eq 0 ]]; then
      echo "postgres did not become ready" >&2
      return 1
    fi
    sleep 1
  done
}

wait_for_http() {
  local url="$1"
  local attempts=90
  until curl --fail --silent --show-error "$url" >/dev/null 2>&1; do
    attempts=$((attempts - 1))
    if [[ "$attempts" -eq 0 ]]; then
      echo "$url did not become ready" >&2
      return 1
    fi
    sleep 1
  done
}

ensure_jwt_key() {
  if [[ -n "${TEST_JWT_PRIVATE_KEY_PATH:-}" && -f "$TEST_JWT_PRIVATE_KEY_PATH" ]]; then
    export TEST_JWT_PRIVATE_KEY_PATH
    return
  fi

  TEST_JWT_PRIVATE_KEY_PATH="/tmp/online-checkers-test-jwt-private.pem"
  openssl genpkey \
    -algorithm RSA \
    -pkeyopt rsa_keygen_bits:2048 \
    -out "$TEST_JWT_PRIVATE_KEY_PATH" >/dev/null 2>&1
  chmod 600 "$TEST_JWT_PRIVATE_KEY_PATH"
  GENERATED_JWT_KEY=1
  export TEST_JWT_PRIVATE_KEY_PATH
}

configure_managed_urls() {
  export E2E_AUTH_URL="${E2E_AUTH_URL:-http://localhost:18081}"
  export E2E_MATCHMAKING_URL="${E2E_MATCHMAKING_URL:-http://localhost:18082}"
  export E2E_GAME_URL="${E2E_GAME_URL:-http://localhost:18083}"
  export E2E_RATING_URL="${E2E_RATING_URL:-http://localhost:18084}"
  export E2E_PROFILE_URL="${E2E_PROFILE_URL:-http://localhost:18086}"
  export E2E_KAFKA_BROKERS="${E2E_KAFKA_BROKERS:-localhost:19092}"
}

start_e2e_stack() {
  if [[ "${TEST_MANAGE_STACK:-1}" == "1" && "$MANAGED_STACK" == "0" ]]; then
    ensure_jwt_key
    "${COMPOSE[@]}" up -d postgres redpanda
    MANAGED_STACK=1
    wait_for_postgres
    run_migrations
    "${COMPOSE[@]}" up -d --build \
      auth-service profile-service rating-service game-service matchmaking-service
    configure_managed_urls
  fi

  wait_for_http "${E2E_AUTH_URL:-http://localhost:8081}/health"
  wait_for_http "${E2E_MATCHMAKING_URL:-http://localhost:8082}/health"
  wait_for_http "${E2E_GAME_URL:-http://localhost:8083}/health"
  wait_for_http "${E2E_RATING_URL:-http://localhost:8084}/health"
  wait_for_http "${E2E_PROFILE_URL:-http://localhost:8086}/health"
}

run_unit() {
  (cd "$ROOT_DIR" && go test ./...)
}

run_integration() {
  if ! psql "$DATABASE_URL" -c "SELECT 1" >/dev/null 2>&1; then
    "${COMPOSE[@]}" up -d postgres
    MANAGED_STACK=1
  fi
  wait_for_postgres
  (cd "$ROOT_DIR" && go test -tags=integration ./services/.../repository)
  if [[ "$MANAGED_STACK" == "1" && "${TEST_KEEP_STACK:-0}" != "1" ]]; then
    "${COMPOSE[@]}" down -v
    MANAGED_STACK=0
  fi
}

run_migrations() {
  (cd "$ROOT_DIR" && \
    goose -table goose_auth_version -dir migrations/auth postgres "$DATABASE_URL" up && \
    goose -table goose_game_version -dir migrations/game postgres "$DATABASE_URL" up && \
    goose -table goose_matchmaking_version -dir migrations/matchmaking postgres "$DATABASE_URL" up && \
    goose -table goose_rating_version -dir migrations/rating postgres "$DATABASE_URL" up && \
    goose -table goose_profile_version -dir migrations/profile postgres "$DATABASE_URL" up)
}

run_go_e2e() {
  start_e2e_stack
  (cd "$ROOT_DIR" && go test -tags=e2e -count=1 ./tests/e2e)
}

postman_env_value() {
  node -e '
    const fs = require("fs");
    const env = JSON.parse(fs.readFileSync(process.argv[1], "utf8"));
    const item = env.values.find((value) => value.key === process.argv[2]);
    if (!item || !item.value) process.exit(2);
    process.stdout.write(String(item.value));
  ' "$POSTMAN_ENV_FILE" "$1"
}

prepare_postman_environment() {
  cp "$ROOT_DIR/postman/ci.postman_environment.json" "$POSTMAN_ENV_FILE"
  node -e '
    const fs = require("fs");
    const filename = process.argv[1];
    const env = JSON.parse(fs.readFileSync(filename, "utf8"));
    const overrides = {
      authUrl: process.env.E2E_AUTH_URL,
      matchmakingUrl: process.env.E2E_MATCHMAKING_URL,
      gameUrl: process.env.E2E_GAME_URL,
      ratingUrl: process.env.E2E_RATING_URL,
      profileUrl: process.env.E2E_PROFILE_URL,
    };
    for (const [key, value] of Object.entries(overrides)) {
      if (!value) continue;
      const item = env.values.find((candidate) => candidate.key === key);
      if (item) item.value = value;
    }
    fs.writeFileSync(filename, JSON.stringify(env, null, 2) + "\n");
  ' "$POSTMAN_ENV_FILE"
}

run_postman() {
  start_e2e_stack
  mkdir -p "$POSTMAN_RESULTS_DIR"
  prepare_postman_environment

  newman run "$ROOT_DIR/postman/identity-profile.postman_collection.json" \
    --environment "$POSTMAN_ENV_FILE" \
    --export-environment "$POSTMAN_ENV_FILE" \
    --bail \
    --reporters cli,junit \
    --reporter-junit-export "$POSTMAN_RESULTS_DIR/identity-profile.xml"

  local user_a_id
  local user_b_id
  user_a_id="$(postman_env_value userAId)"
  user_b_id="$(postman_env_value userBId)"

  (cd "$ROOT_DIR" && go run -tags=e2e ./tests/e2e/fixtures \
    -brokers "${E2E_KAFKA_BROKERS:-localhost:9092}" \
    -white "$user_a_id" \
    -black "$user_b_id")

  newman run "$ROOT_DIR/postman/rating-matchmaking-game.postman_collection.json" \
    --environment "$POSTMAN_ENV_FILE" \
    --export-environment "$POSTMAN_ENV_FILE" \
    --bail \
    --reporters cli,junit \
    --reporter-junit-export "$POSTMAN_RESULTS_DIR/rating-matchmaking-game.xml"
}

run_e2e() {
  run_go_e2e
  run_postman
}

case "$MODE" in
  unit)
    run_unit
    ;;
  integration)
    run_integration
    ;;
  e2e)
    run_e2e
    ;;
  postman)
    run_postman
    ;;
  all)
    run_unit
    run_integration
    run_e2e
    ;;
  -h|--help|help)
    usage
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac
