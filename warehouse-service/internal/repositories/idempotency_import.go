package repositories

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/industrial-sed/warehouse-service/internal/models"
)

// GetIdempotency возвращает сохранённый ответ.
func (s *Store) GetIdempotency(ctx context.Context, tx pgx.Tx, tenant, key string) ([]byte, error) {
	var raw []byte
	err := s.db(tx).QueryRow(ctx, `SELECT response FROM idempotency_keys WHERE tenant_code = $1 AND key = $2`, tenant, key).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return raw, err
}

// SaveIdempotency сохраняет ответ.
func (s *Store) SaveIdempotency(ctx context.Context, tx pgx.Tx, tenant, key, hash string, response interface{}) error {
	b, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, err = s.db(tx).Exec(ctx, `
		INSERT INTO idempotency_keys (key, tenant_code, request_hash, response) VALUES ($1,$2,$3,$4)
		ON CONFLICT (tenant_code, key) DO NOTHING
	`, key, tenant, hash, b)
	return err
}

// CreateImportJob создаёт задачу импорта.
func (s *Store) CreateImportJob(ctx context.Context, tx pgx.Tx, j *models.ImportJob) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO import_jobs (id, tenant_code, kind, status, total, processed, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, j.ID, j.TenantCode, j.Kind, j.Status, j.Total, j.Processed, j.CreatedBy)
	return err
}

// UpdateImportJob обновляет прогресс.
func (s *Store) UpdateImportJob(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string, processed int, errs []byte) error {
	_, err := s.db(tx).Exec(ctx, `
		UPDATE import_jobs SET status = $2, processed = $3, errors = $4, finished_at = CASE WHEN $2 IN ('DONE','FAILED') THEN now() ELSE finished_at END
		WHERE id = $1
	`, id, status, processed, errs)
	return err
}

// GetImportJob задача.
func (s *Store) GetImportJob(ctx context.Context, tx pgx.Tx, tenant string, id uuid.UUID) (*models.ImportJob, error) {
	var j models.ImportJob
	var errB []byte
	err := s.db(tx).QueryRow(ctx, `
		SELECT id, tenant_code, kind, status, total, processed, errors, created_by, created_at, finished_at
		FROM import_jobs WHERE id = $1 AND tenant_code = $2
	`, id, tenant).Scan(&j.ID, &j.TenantCode, &j.Kind, &j.Status, &j.Total, &j.Processed, &errB, &j.CreatedBy, &j.CreatedAt, &j.FinishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	j.Errors = errB
	return &j, err
}
