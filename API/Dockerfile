# Etapa 1: compilación
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copiamos solo el go.mod para evitar errores si go.sum no existe aún
COPY go.mod ./

# Si ya tenés un archivo go.sum, descomentá la línea de abajo:
# COPY go.sum ./

RUN go mod download
COPY . .
RUN go get github.com/lib/pq
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
