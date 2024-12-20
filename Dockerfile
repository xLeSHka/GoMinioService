FROM golang:1.23-alpine AS builder

RUN apk --no-cache add ca-certificates gcc g++ libc-dev

WORKDIR /app

COPY ../go.mod ../go.sum ./

COPY . .
RUN CGO_ENABLED=0 go build -o file_service ./cmd/main
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/file_service /file_service
COPY --from=builder /app/config.env /config.env

CMD [ "/file_service" ]