FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /auth-service ./services/auth/cmd

FROM alpine:3.20
COPY --from=builder /auth-service /auth-service

EXPOSE 8081

CMD ["/auth-service"]