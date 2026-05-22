package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/industrial-sed/sed-service/internal/models"
	"github.com/industrial-sed/sed-service/internal/ragtext"
)

// RagExportPermissions кто может читать / писать / согласовывать документ этого типа.
type RagExportPermissions struct {
	ReadRoles    []string `json:"read_roles"`
	WriteRoles   []string `json:"write_roles"`
	ApproveRoles []string `json:"approve_roles"`
	AdminRoles   []string `json:"admin_roles"`
}

// RagExportAttachment ссылка на бинарное вложение (скачивание с тем же X-Service-Secret).
type RagExportAttachment struct {
	FileID      string  `json:"file_id"`
	Name        string  `json:"name"`
	ContentType *string `json:"content_type,omitempty"`
	SizeBytes   int64   `json:"size_bytes"`
	URL         string  `json:"url"`
}

// RagExportDocument одна карточка для выгрузки в RAG.
type RagExportDocument struct {
	DocumentID  string                `json:"document_id"`
	Text        string                `json:"text"`
	Access      RagExportPermissions  `json:"access"`
	Attachments []RagExportAttachment `json:"attachments"`
}

// RagExportResponse минимальный корпус: только текст, ACL и файлы.
type RagExportResponse struct {
	Documents []RagExportDocument `json:"documents"`
}

var (
	ragApproveRoles = []string{models.RoleSedApprover}
	ragAdminRoles   = []string{models.RoleSedAdmin}
)

// BuildRagExport корпус документов тенанта для внешнего RAG-модуля.
func (a *App) BuildRagExport(ctx context.Context, tenant, publicBaseURL string) (*RagExportResponse, error) {
	rows, err := a.Store.ListRagCorpusDocuments(ctx, tenant)
	if err != nil {
		return nil, err
	}
	fileRows, err := a.Store.ListFilesByTenant(ctx, tenant)
	if err != nil {
		return nil, err
	}
	filesByDoc := map[string][]RagExportAttachment{}
	base := strings.TrimRight(publicBaseURL, "/")
	for _, f := range fileRows {
		url := fmt.Sprintf("%s/api/v1/internal/rag/documents/%s/files/%s", base, f.DocumentID, f.FileID)
		filesByDoc[f.DocumentID] = append(filesByDoc[f.DocumentID], RagExportAttachment{
			FileID: f.FileID, Name: f.OriginalName, ContentType: f.ContentType,
			SizeBytes: f.SizeBytes, URL: url,
		})
	}

	docs := make([]RagExportDocument, 0, len(rows))
	for _, r := range rows {
		text := ragtext.FormatDocumentText(r.Title, json.RawMessage(r.Payload))
		atts := filesByDoc[r.ID]
		if atts == nil {
			atts = []RagExportAttachment{}
		}
		docs = append(docs, RagExportDocument{
			DocumentID: r.ID,
			Text:       text,
			Access: RagExportPermissions{
				ReadRoles:    append([]string(nil), r.ReaderRoles...),
				WriteRoles:   append([]string(nil), r.WriterRoles...),
				ApproveRoles: append([]string(nil), ragApproveRoles...),
				AdminRoles:   append([]string(nil), ragAdminRoles...),
			},
			Attachments: atts,
		})
	}
	return &RagExportResponse{Documents: docs}, nil
}

// OpenDocumentFileInternal скачивание вложения без JWT (только internal RAG).
func (a *App) OpenDocumentFileInternal(ctx context.Context, tenant string, docID, fileID uuid.UUID) (*models.DocumentFile, io.ReadCloser, error) {
	if a.Minio == nil {
		return nil, nil, ErrValidation
	}
	d, err := a.Store.GetDocument(ctx, nil, tenant, docID)
	if err != nil {
		return nil, nil, err
	}
	if d == nil {
		return nil, nil, ErrNotFound
	}
	meta, err := a.Store.GetDocumentFile(ctx, nil, tenant, docID, fileID)
	if err != nil {
		return nil, nil, err
	}
	if meta == nil {
		return nil, nil, ErrNotFound
	}
	rc, err := a.OpenFileStream(ctx, meta.ObjectKey)
	if err != nil {
		return nil, nil, err
	}
	return meta, rc, nil
}

// UpdateDocumentFixture обновляет title/payload (rag_content опционален, для seed).
func (a *App) UpdateDocumentFixture(ctx context.Context, tenant string, docID uuid.UUID, title string, payload, rag json.RawMessage) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	d, err := a.Store.LockDocumentForUpdate(ctx, tx, tenant, docID)
	if err != nil {
		return err
	}
	if d == nil {
		return ErrNotFound
	}
	d.Title = title
	if len(payload) > 0 {
		d.Payload = payload
	}
	if err := a.Store.UpdateDocument(ctx, tx, d); err != nil {
		return err
	}
	if len(rag) > 0 {
		if err := a.Store.UpdateDocumentRagContent(ctx, tx, tenant, docID, rag); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// SetDocumentRagContent сохраняет rag_content (legacy seed).
func (a *App) SetDocumentRagContent(ctx context.Context, tenant string, docID uuid.UUID, rag json.RawMessage) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	d, err := a.Store.GetDocument(ctx, tx, tenant, docID)
	if err != nil {
		return err
	}
	if d == nil {
		return ErrNotFound
	}
	if err := a.Store.UpdateDocumentRagContent(ctx, tx, tenant, docID, rag); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
