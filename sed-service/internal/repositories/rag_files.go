package repositories

import (
	"context"
)

// TenantFileRow вложение документа тенанта.
type TenantFileRow struct {
	FileID       string
	DocumentID   string
	OriginalName string
	ContentType  *string
	SizeBytes    int64
	ObjectKey    string
}

// ListFilesByTenant все файлы документов тенанта.
func (s *Store) ListFilesByTenant(ctx context.Context, tenant string) ([]TenantFileRow, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT f.id::text, f.document_id::text, f.original_name, f.content_type, f.size_bytes, f.object_key
		FROM document_files f
		JOIN documents d ON d.id = f.document_id
		WHERE d.tenant_code = $1
		ORDER BY f.document_id, f.uploaded_at
	`, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TenantFileRow
	for rows.Next() {
		var r TenantFileRow
		if err := rows.Scan(&r.FileID, &r.DocumentID, &r.OriginalName, &r.ContentType, &r.SizeBytes, &r.ObjectKey); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
