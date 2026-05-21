FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY platform ./platform
COPY production-service/go.mod production-service/go.sum ./production-service/
WORKDIR /src/production-service
RUN go mod download
COPY production-service .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/production-service ./cmd/production-service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /out/production-service /app/production-service
COPY production-service/migrations ./migrations
COPY production-service/configs ./configs
EXPOSE 8092
HEALTHCHECK --interval=10s --timeout=3s --start-period=40s --retries=5 \
  CMD wget -q -O- http://127.0.0.1:8092/health || exit 1
ENTRYPOINT ["/app/production-service"]
