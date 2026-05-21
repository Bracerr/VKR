FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY platform ./platform
COPY sed-service/go.mod sed-service/go.sum ./sed-service/
WORKDIR /src/sed-service
RUN go mod download
COPY sed-service .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/sed-service ./cmd/sed-service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /out/sed-service /app/sed-service
COPY sed-service/migrations ./migrations
COPY sed-service/configs ./configs
EXPOSE 8091
HEALTHCHECK --interval=10s --timeout=3s --start-period=40s --retries=5 \
  CMD wget -q -O- http://127.0.0.1:8091/health || exit 1
ENTRYPOINT ["/app/sed-service"]
