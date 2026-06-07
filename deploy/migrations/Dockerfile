FROM golang:1.25-alpine AS builder

RUN go install \
    -tags="no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb" \
    github.com/pressly/goose/v3/cmd/goose@v3.27.1

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations /migrations
COPY deploy/migrations/run.sh /usr/local/bin/run-migrations

RUN chmod +x /usr/local/bin/run-migrations

ENTRYPOINT ["/usr/local/bin/run-migrations"]
