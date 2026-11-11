FROM sqlc/sqlc:1.31.1 AS sqlc
FROM ghcr.io/a-h/templ:v0.3.1100 AS templ

FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY --from=sqlc /workspace/sqlc /usr/local/bin/sqlc
COPY --from=templ /bin/templ /usr/local/bin/templ
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN sqlc generate && templ generate && CGO_ENABLED=0 go build -o /app/bookfly ./API

FROM alpine:3.20
RUN adduser -D -u 1000 app
USER app
WORKDIR /app
COPY --from=builder /app/bookfly /usr/local/bin/bookfly
COPY --from=builder /app/API/static ./API/static
EXPOSE 8080
CMD ["bookfly"]
