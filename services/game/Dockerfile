FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /game-service ./services/game/cmd

FROM alpine:3.20
COPY --from=builder /game-service /game-service

EXPOSE 8081

CMD ["/game-service"]