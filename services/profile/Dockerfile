FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /profile-service ./services/profile/cmd

FROM alpine:3.20
COPY --from=builder /profile-service /profile-service

EXPOSE 8081

CMD ["/profile-service"]