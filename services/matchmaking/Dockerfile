FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /matchmaking-service ./services/matchmaking/cmd

FROM alpine:3.20
COPY --from=builder /matchmaking-service /matchmaking-service

EXPOSE 8081

CMD ["/matchmaking-service"]