#!/bin/sh

set -eu

: "${DATABASE_URL:?DATABASE_URL is required}"

run_migration() {
  service="$1"
  goose \
    -table "goose_${service}_version" \
    -dir "/migrations/${service}" \
    postgres \
    "$DATABASE_URL" \
    up
}

run_migration auth
run_migration game
run_migration matchmaking
run_migration rating
run_migration profile
