FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/warehouse-service ./cmd/warehouse-service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /out/warehouse-service /app/warehouse-service
COPY migrations ./migrations
COPY configs ./configs
EXPOSE 8090
HEALTHCHECK --interval=10s --timeout=3s --start-period=40s --retries=5 \
  CMD wget -q -O- http://127.0.0.1:8090/health || exit 1
ENTRYPOINT ["/app/warehouse-service"]
