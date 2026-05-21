package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/industrial-sed/platform/events"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusDelivered  = "delivered"
	StatusFailed     = "failed"
)

// Row запись outbox.
type Row struct {
	ID           uuid.UUID
	Topic        string
	MessageKey   string
	Payload      []byte
	Headers      map[string]string
	Status       string
	Attempts     int
	NextRetryAt  time.Time
	LastError    string
	CreatedAt    time.Time
}

// Store работа с outbox_events.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Enqueue добавляет событие в outbox (отдельная транзакция).
func (s *Store) Enqueue(ctx context.Context, topic, messageKey string, env events.Envelope) error {
	body, err := env.Marshal()
	if err != nil {
		return err
	}
	headers, err := json.Marshal(env.Headers())
	if err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (topic, message_key, payload, headers, status, next_retry_at)
		VALUES ($1, $2, $3, $4, $5, now())`,
		topic, messageKey, body, headers, StatusPending,
	)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// EnqueueTx в той же транзакции, что бизнес-операция.
func (s *Store) EnqueueTx(ctx context.Context, tx pgx.Tx, topic, messageKey string, env events.Envelope) error {
	body, err := env.Marshal()
	if err != nil {
		return err
	}
	headers, err := json.Marshal(env.Headers())
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (topic, message_key, payload, headers, status, next_retry_at)
		VALUES ($1, $2, $3, $4, $5, now())`,
		topic, messageKey, body, headers, StatusPending,
	)
	return err
}

// FetchPendingBatch выбирает pending для relay.
func (s *Store) FetchPendingBatch(ctx context.Context, limit int) ([]Row, error) {
	if limit <= 0 {
		limit = 50
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id, topic, message_key, payload, headers, status, attempts, next_retry_at, COALESCE(last_error,''), created_at
		FROM outbox_events
		WHERE status IN ($1, $2) AND next_retry_at <= now()
		ORDER BY created_at
		LIMIT $3
		FOR UPDATE SKIP LOCKED`,
		StatusPending, StatusFailed, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Row
	for rows.Next() {
		var r Row
		var hdr []byte
		if err := rows.Scan(&r.ID, &r.Topic, &r.MessageKey, &r.Payload, &hdr, &r.Status, &r.Attempts, &r.NextRetryAt, &r.LastError, &r.CreatedAt); err != nil {
			return nil, err
		}
		if len(hdr) > 0 {
			_ = json.Unmarshal(hdr, &r.Headers)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		_, err := tx.Exec(ctx, `UPDATE outbox_events SET status = $1 WHERE id = $2`, StatusProcessing, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Status = StatusProcessing
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) MarkDelivered(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `UPDATE outbox_events SET status = $1, last_error = NULL WHERE id = $2`, StatusDelivered, id)
	return err
}

func (s *Store) MarkFailed(ctx context.Context, id uuid.UUID, errMsg string, attempts int, nextRetry time.Time) error {
	status := StatusFailed
	if attempts >= maxAttempts {
		status = StatusFailed
	}
	_, err := s.pool.Exec(ctx, `
		UPDATE outbox_events SET status = $1, attempts = $2, last_error = $3, next_retry_at = $4
		WHERE id = $5`,
		status, attempts, errMsg, nextRetry, id,
	)
	return err
}

func (s *Store) ResetToPending(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE outbox_events SET status = $1, next_retry_at = now(), last_error = NULL WHERE id = $2`,
		StatusPending, id,
	)
	return err
}

// ListFailed для admin replay.
func (s *Store) ListFailed(ctx context.Context, limit int) ([]Row, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, topic, message_key, payload, headers, status, attempts, next_retry_at, COALESCE(last_error,''), created_at
		FROM outbox_events WHERE status = $1 ORDER BY created_at DESC LIMIT $2`, StatusFailed, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Row
	for rows.Next() {
		var r Row
		var hdr []byte
		if err := rows.Scan(&r.ID, &r.Topic, &r.MessageKey, &r.Payload, &hdr, &r.Status, &r.Attempts, &r.NextRetryAt, &r.LastError, &r.CreatedAt); err != nil {
			return nil, err
		}
		if len(hdr) > 0 {
			_ = json.Unmarshal(hdr, &r.Headers)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

const maxAttempts = 10

func Backoff(attempts int) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	sec := 1 << min(attempts, 6)
	if sec > 60 {
		sec = 60
	}
	return time.Duration(sec) * time.Second
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// EnvelopeFromRow декодирует payload.
func EnvelopeFromRow(r Row) (events.Envelope, error) {
	var e events.Envelope
	if err := json.Unmarshal(r.Payload, &e); err != nil {
		return events.Envelope{}, fmt.Errorf("decode envelope: %w", err)
	}
	return e, nil
}
