FROM golang:1.23.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o file_service ./cmd/main

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/file_service /app/file_service

CMD /app/file_service