# Сборка auth-service (multi-stage).
FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/auth-service ./cmd/auth-service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /out/auth-service /app/auth-service
COPY migrations ./migrations
COPY config.yaml ./config.yaml
COPY configs ./configs
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=40s --retries=5 \
  CMD wget -q -O- http://127.0.0.1:8080/health || exit 1
ENTRYPOINT ["/app/auth-service"]
