# Etapa 1: compilación
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copiamos todo el código fuente de una vez para que Go tenga el contexto completo
COPY . .

# Descargamos las dependencias y el driver de Postgres directamente
RUN go mod tidy
RUN go get github.com/lib/pq

# Compilamos la aplicación de Go
RUN CGO_ENABLED=0 go build -o /app/api .

# Etapa 2: imagen final
FROM alpine:3.20
RUN adduser -D -u 1000 app
USER app
WORKDIR /app
COPY --from=builder /app/api /usr/local/bin/api
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["api"]
