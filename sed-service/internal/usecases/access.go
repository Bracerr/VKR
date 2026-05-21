package usecases

import (
	"context"

	"github.com/google/uuid"

	"github.com/industrial-sed/sed-service/internal/models"
)

func (a *App) ensureDocumentRead(ctx context.Context, tenant string, docID uuid.UUID, c *models.Claims) (*models.Document, error) {
	d, err := a.Store.GetDocument(ctx, nil, tenant, docID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, ErrNotFound
	}
	if models.CanAdminSED(c) {
		return d, nil
	}
	dt, err := a.Store.GetDocumentType(ctx, nil, tenant, d.TypeID)
	if err != nil {
		return nil, err
	}
	if dt == nil {
		return nil, ErrNotFound
	}
	pending := false
	if models.CanApprove(c) {
		pending, err = a.Store.DocumentHasPendingApprovalForUser(ctx, nil, tenant, docID, c.Sub, c.RealmRoles)
		if err != nil {
			return nil, err
		}
	}
	if !models.CanReadDocument(c, d, dt, pending) {
		return nil, ErrForbidden
	}
	return d, nil
}

func resolveTypeACL(code, whAction string, readers, writers []string) ([]string, []string) {
	if len(readers) == 0 && len(writers) == 0 {
		return models.DefaultACLForDocumentType(code, whAction)
	}
	if len(readers) == 0 {
		r, _ := models.DefaultACLForDocumentType(code, whAction)
		readers = r
	}
	if len(writers) == 0 {
		_, w := models.DefaultACLForDocumentType(code, whAction)
		writers = w
	}
	if readers == nil {
		readers = []string{}
	}
	if writers == nil {
		writers = []string{}
	}
	return readers, writers
}
