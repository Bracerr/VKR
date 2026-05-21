FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY platform ./platform
COPY traceability-service/go.mod traceability-service/go.sum ./traceability-service/
WORKDIR /src/traceability-service
RUN go mod download
COPY traceability-service .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/traceability-service ./cmd/traceability-service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /out/traceability-service /app/traceability-service
COPY traceability-service/migrations ./migrations
COPY traceability-service/configs ./configs
EXPOSE 8095
HEALTHCHECK --interval=10s --timeout=3s --start-period=40s --retries=5 \
  CMD wget -q -O- http://127.0.0.1:8095/health || exit 1
ENTRYPOINT ["/app/traceability-service"]
