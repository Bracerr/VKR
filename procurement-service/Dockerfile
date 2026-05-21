FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY platform ./platform
COPY procurement-service/go.mod procurement-service/go.sum ./procurement-service/
WORKDIR /src/procurement-service
RUN go mod download
COPY procurement-service .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/procurement-service ./cmd/procurement-service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /out/procurement-service /app/procurement-service
COPY procurement-service/migrations ./migrations
COPY procurement-service/configs ./configs
EXPOSE 8093
HEALTHCHECK --interval=10s --timeout=3s --start-period=40s --retries=5 \
  CMD wget -q -O- http://127.0.0.1:8093/health || exit 1
ENTRYPOINT ["/app/procurement-service"]
