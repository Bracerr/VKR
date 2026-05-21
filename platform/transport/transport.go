package transport

import "strings"

const (
	HTTP  = "http"
	Kafka = "kafka"
	Dual  = "dual"
)

// Mode нормализует EVENT_TRANSPORT.
func Mode(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case Kafka, Dual:
		return strings.ToLower(strings.TrimSpace(s))
	default:
		return HTTP
	}
}

func UseKafka(mode string) bool {
	m := Mode(mode)
	return m == Kafka || m == Dual
}

func UseHTTP(mode string) bool {
	m := Mode(mode)
	return m == HTTP || m == Dual
}
