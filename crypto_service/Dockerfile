FROM golang:1.23.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o crypto_service ./cmd/main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/crypto_service /app/crypto_service

ENTRYPOINT ["./crypto_service"]