FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /sed-service ./cmd/sed-service

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=build /sed-service /app/sed-service
COPY configs /app/configs
COPY migrations /app/migrations
ENV CONFIG_PATH=/app/configs/docker.yaml
EXPOSE 8091
HEALTHCHECK --interval=10s --timeout=3s --start-period=30s --retries=5 \
  CMD wget -q -O- http://127.0.0.1:8091/health || exit 1
ENTRYPOINT ["/app/sed-service"]
