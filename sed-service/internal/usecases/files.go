package usecases

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/industrial-sed/sed-service/internal/clients"
	"github.com/industrial-sed/sed-service/internal/models"
)

// UploadFile загружает вложение в MinIO и БД (только черновик, автор).
func (a *App) UploadFile(ctx context.Context, tenant, uploader string, docID uuid.UUID, origName, contentType string, size int64, r io.Reader) (*models.DocumentFile, error) {
	if a.Minio == nil {
		return nil, ErrValidation
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	d, err := a.Store.GetDocument(ctx, tx, tenant, docID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, ErrNotFound
	}
	if d.Status != "DRAFT" {
		return nil, ErrWrongState
	}
	if d.AuthorSub != uploader {
		return nil, ErrForbidden
	}
	fid := uuid.New()
	key := clients.ObjectKey(tenant, docID.String(), fid.String(), origName)
	if err := a.Minio.Put(ctx, key, r, size, contentType); err != nil {
		return nil, err
	}
	ct := contentType
	f := &models.DocumentFile{
		ID: fid, DocumentID: docID, ObjectKey: key, OriginalName: origName,
		ContentType: &ct, SizeBytes: size, UploadedBy: uploader,
	}
	if err := a.Store.InsertDocumentFile(ctx, tx, f); err != nil {
		_ = a.Minio.Remove(ctx, key)
		return nil, err
	}
	if err := a.Store.InsertHistory(ctx, tx, docID, uploader, "FILE_UPLOAD", histPayload(map[string]any{"file_id": fid.String()})); err != nil {
		_ = a.Minio.Remove(ctx, key)
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		_ = a.Minio.Remove(ctx, key)
		return nil, err
	}
	return f, nil
}

// GetFileMeta метаданные файла (проверка тенанта и документа).
func (a *App) GetFileMeta(ctx context.Context, tenant string, docID, fileID uuid.UUID) (*models.DocumentFile, error) {
	f, err := a.Store.GetDocumentFile(ctx, nil, tenant, docID, fileID)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, ErrNotFound
	}
	return f, nil
}

// OpenFileStream открывает объект в MinIO (после GetFileMeta / права в handler).
func (a *App) OpenFileStream(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	if a.Minio == nil {
		return nil, ErrValidation
	}
	o, err := a.Minio.Get(ctx, objectKey)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// ListDocumentFiles список вложений.
func (a *App) ListDocumentFiles(ctx context.Context, tenant string, docID uuid.UUID) ([]models.DocumentFile, error) {
	d, err := a.Store.GetDocument(ctx, nil, tenant, docID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, ErrNotFound
	}
	return a.Store.ListDocumentFiles(ctx, nil, docID)
}

// DeleteFile удаляет вложение (черновик, автор).
func (a *App) DeleteFile(ctx context.Context, tenant, actor string, docID, fileID uuid.UUID) error {
	if a.Minio == nil {
		return ErrValidation
	}
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
	if d.Status != "DRAFT" {
		return ErrWrongState
	}
	if d.AuthorSub != actor {
		return ErrForbidden
	}
	f, err := a.Store.GetDocumentFile(ctx, tx, tenant, docID, fileID)
	if err != nil {
		return err
	}
	if f == nil {
		return ErrNotFound
	}
	if err := a.Store.DeleteDocumentFile(ctx, tx, fileID); err != nil {
		return err
	}
	if err := a.Store.InsertHistory(ctx, tx, docID, actor, "FILE_DELETE", histPayload(map[string]any{"file_id": fileID.String()})); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	_ = a.Minio.Remove(ctx, f.ObjectKey)
	return nil
}
