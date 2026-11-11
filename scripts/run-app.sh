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
direccion=$(docker compose port api 8080)
printf '\nBookfly disponible en http://%s (Ctrl+C para salir y limpiar).\n\n' "$direccion"
docker compose logs --follow api
