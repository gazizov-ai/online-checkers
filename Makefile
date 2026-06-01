COMPOSE = docker compose -f deploy/docker-compose.yml
DATABASE_URL = postgresql://postgres:postgres@localhost:5432/checkers?sslmode=disable

.PHONY: compose-up compose-up-d compose-down compose-down-v compose-logs migrate-up migrate-status migrate-down

compose-up:
	$(COMPOSE) up --build

compose-up-d:
	$(COMPOSE) up -d --build

compose-down:
	$(COMPOSE) down

compose-down-v:
	$(COMPOSE) down -v

compose-logs:
	$(COMPOSE) logs -f

migrate-up:
	goose -table goose_auth_version -dir migrations/auth postgres "$(DATABASE_URL)" up
	goose -table goose_game_version -dir migrations/game postgres "$(DATABASE_URL)" up
	goose -table goose_matchmaking_version -dir migrations/matchmaking postgres "$(DATABASE_URL)" up
	goose -table goose_rating_version -dir migrations/rating postgres "$(DATABASE_URL)" up

migrate-status:
	goose -table goose_auth_version -dir migrations/auth postgres "$(DATABASE_URL)" status
	goose -table goose_game_version -dir migrations/game postgres "$(DATABASE_URL)" status
	goose -table goose_matchmaking_version -dir migrations/matchmaking postgres "$(DATABASE_URL)" status
	goose -table goose_rating_version -dir migrations/rating postgres "$(DATABASE_URL)" status

migrate-down:
	goose -table goose_rating_version -dir migrations/rating postgres "$(DATABASE_URL)" down
	goose -table goose_matchmaking_version -dir migrations/matchmaking postgres "$(DATABASE_URL)" down
	goose -table goose_game_version -dir migrations/game postgres "$(DATABASE_URL)" down
	goose -table goose_auth_version -dir migrations/auth postgres "$(DATABASE_URL)" down