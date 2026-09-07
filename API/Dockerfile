# Etapa 1: compilación
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN go get github.com/lib/pq
RUN CGO_ENABLED=0 go build -o /app/api .

# Etapa 2: imagen final
FROM alpine:3.20
RUN adduser -D -u 1000 app
USER app
# Se establece un directorio de trabajo para la ejecución
WORKDIR /app
COPY --from=builder /app/api /usr/local/bin/api
# Se copia la carpeta estática para que el servidor la encuentre
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["api"]
