# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/texa-api ./cmd/server

FROM alpine:3.21
WORKDIR /app
RUN adduser -D -u 10001 appuser

COPY --from=builder /bin/texa-api /app/texa-api
COPY backend/migrations /app/migrations

USER appuser
EXPOSE 8081
CMD ["/app/texa-api"]
