# Etapa de compilación
FROM golang:1.26-alpine AS builder
WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar el binario estático (ejemplo de servicio que usa la librería)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/make-cfdi

# Etapa final de ejecución
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copiar el binario compilado y assets si tu app los necesita en ejecución
COPY --from=builder /app/main .
COPY --from=builder /app/assets ./assets

EXPOSE 8080
CMD ["./main"]
