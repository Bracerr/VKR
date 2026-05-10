package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/industrial-sed/sed-service/internal/models"
)

// InsertDocumentFile метаданные файла.
func (s *Store) InsertDocumentFile(ctx context.Context, tx pgx.Tx, f *models.DocumentFile) error {
	_, err := s.db(tx).Exec(ctx, `
		INSERT INTO document_files (id, document_id, object_key, original_name, content_type, size_bytes, uploaded_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, f.ID, f.DocumentID, f.ObjectKey, f.OriginalName, f.ContentType, f.SizeBytes, f.UploadedBy)
	return err
}

// GetDocumentFile файл.
func (s *Store) GetDocumentFile(ctx context.Context, tx pgx.Tx, tenant string, docID, fileID uuid.UUID) (*models.DocumentFile, error) {
	var f models.DocumentFile
	err := s.db(tx).QueryRow(ctx, `
		SELECT f.id, f.document_id, f.object_key, f.original_name, f.content_type, f.size_bytes, f.uploaded_by, f.uploaded_at
		FROM document_files f
		JOIN documents d ON d.id = f.document_id
		WHERE f.id = $1 AND f.document_id = $2 AND d.tenant_code = $3
	`, fileID, docID, tenant).Scan(
		&f.ID, &f.DocumentID, &f.ObjectKey, &f.OriginalName, &f.ContentType, &f.SizeBytes, &f.UploadedBy, &f.UploadedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &f, err
}

// ListDocumentFiles список.
func (s *Store) ListDocumentFiles(ctx context.Context, tx pgx.Tx, docID uuid.UUID) ([]models.DocumentFile, error) {
	rows, err := s.db(tx).Query(ctx, `
		SELECT id, document_id, object_key, original_name, content_type, size_bytes, uploaded_by, uploaded_at
		FROM document_files WHERE document_id = $1 ORDER BY uploaded_at
	`, docID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.DocumentFile
	for rows.Next() {
		var f models.DocumentFile
		if err := rows.Scan(&f.ID, &f.DocumentID, &f.ObjectKey, &f.OriginalName, &f.ContentType, &f.SizeBytes, &f.UploadedBy, &f.UploadedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// DeleteDocumentFile удаление метаданных.
func (s *Store) DeleteDocumentFile(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	tag, err := s.db(tx).Exec(ctx, `DELETE FROM document_files WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
