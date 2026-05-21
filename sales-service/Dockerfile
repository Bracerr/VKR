FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY platform ./platform
COPY sales-service/go.mod sales-service/go.sum ./sales-service/
WORKDIR /src/sales-service
RUN go mod download
COPY sales-service .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/sales-service ./cmd/sales-service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /out/sales-service /app/sales-service
COPY sales-service/migrations ./migrations
COPY sales-service/configs ./configs
EXPOSE 8094
HEALTHCHECK --interval=10s --timeout=3s --start-period=40s --retries=5 \
  CMD wget -q -O- http://127.0.0.1:8094/health || exit 1
ENTRYPOINT ["/app/sales-service"]
