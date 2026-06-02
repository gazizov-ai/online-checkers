FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /rating-service ./services/rating/cmd

FROM alpine:3.20
COPY --from=builder /rating-service /rating-service

EXPOSE 8081

CMD ["/rating-service"]