package repositories

import (
	"context"
)

// RagCorpusRow документ с метаданными типа для выгрузки RAG.
type RagCorpusRow struct {
	ID          string
	TypeID      string
	TypeCode    string
	TypeName    string
	Number      string
	Title       string
	Status      string
	AuthorSub   string
	Payload     []byte
	RagContent  []byte
	ReaderRoles []string
	WriterRoles []string
}

// ListRagCorpusDocuments все документы тенанта с типами (без ACL).
func (s *Store) ListRagCorpusDocuments(ctx context.Context, tenant string) ([]RagCorpusRow, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT d.id::text, d.type_id::text, dt.code, dt.name, d.number, d.title, d.status, d.author_sub,
		       d.payload, d.rag_content, dt.reader_roles, dt.writer_roles
		FROM documents d
		JOIN document_types dt ON dt.id = d.type_id AND dt.tenant_code = d.tenant_code
		WHERE d.tenant_code = $1
		ORDER BY d.created_at
	`, tenant)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RagCorpusRow
	for rows.Next() {
		var r RagCorpusRow
		if err := rows.Scan(
			&r.ID, &r.TypeID, &r.TypeCode, &r.TypeName, &r.Number, &r.Title, &r.Status, &r.AuthorSub,
			&r.Payload, &r.RagContent, &r.ReaderRoles, &r.WriterRoles,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
