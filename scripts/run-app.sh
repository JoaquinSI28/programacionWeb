#!/usr/bin/env bash
set -eu

cleanup() {
  docker compose --profile test down -v --remove-orphans
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

cleanup
make up
printf '\nBookfly disponible en http://localhost:%s (Ctrl+C para salir y limpiar).\n\n' "${APP_PORT:-8080}"
docker compose logs --follow api
