package config

import (
	"os"
	"strings"

	"github.com/segmentio/kafka-go"
)

// ParseBrokers из KAFKA_BROKERS (comma-separated) или слайса.
func ParseBrokers(envKey string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(envKey))
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func EventTransport() string {
	return strings.TrimSpace(os.Getenv("EVENT_TRANSPORT"))
}

func KafkaConsumerStartOffset() int64 {
	if strings.EqualFold(os.Getenv("KAFKA_CONSUMER_OFFSET"), "earliest") {
		return kafka.FirstOffset
	}
	return kafka.LastOffset
}
