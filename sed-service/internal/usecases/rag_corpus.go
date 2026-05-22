package usecases

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/industrial-sed/sed-service/internal/clients"
	"github.com/industrial-sed/sed-service/internal/models"
	"github.com/industrial-sed/sed-service/internal/repositories"
)

// RagCorpusDocument запись для индексации RAG.
type RagCorpusDocument struct {
	DocumentID  string          `json:"document_id"`
	TypeID      string          `json:"type_id"`
	TypeCode    string          `json:"type_code"`
	TypeName    string          `json:"type_name"`
	Number      string          `json:"number"`
	Title       string          `json:"title"`
	Status      string          `json:"status"`
	AuthorSub   string          `json:"author_sub"`
	Payload     json.RawMessage `json:"payload"`
	Content     json.RawMessage `json:"content"`
	SearchText  string          `json:"search_text"`
	ReaderRoles []string        `json:"reader_roles"`
	WriterRoles []string        `json:"writer_roles"`
}

// RagCorpusUserAccess видимость документов для пользователя.
type RagCorpusUserAccess struct {
	Login               string   `json:"login"`
	Username            string   `json:"username"`
	KeycloakID          string   `json:"keycloak_id"`
	Roles               []string `json:"roles"`
	VisibleDocumentIDs  []string `json:"visible_document_ids"`
	ExpectedCount       int      `json:"expected_count"`
}

// RagCorpusResponse выгрузка корпуса для RAG-модуля.
type RagCorpusResponse struct {
	Tenant    string                `json:"tenant"`
	FetchedAt time.Time             `json:"fetched_at"`
	Documents []RagCorpusDocument   `json:"documents"`
	Users     []RagCorpusUserAccess `json:"users"`
}

// UpdateDocumentFixture обновляет title/payload/rag_content без ACL (только internal seed).
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
	if err := a.Store.UpdateDocumentRagContent(ctx, tx, tenant, docID, rag); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SetDocumentRagContent сохраняет rag_content (internal).
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

// BuildRagCorpus все документы тенанта + матрица видимости по пользователям.
func (a *App) BuildRagCorpus(ctx context.Context, tenant string, authUsers *clients.AuthUsersClient) (*RagCorpusResponse, error) {
	rows, err := a.Store.ListRagCorpusDocuments(ctx, tenant)
	if err != nil {
		return nil, err
	}
	docs := make([]RagCorpusDocument, 0, len(rows))
	for _, r := range rows {
		content := json.RawMessage(r.RagContent)
		if len(content) == 0 {
			content = json.RawMessage(`{}`)
		}
		searchText := extractSearchText(content)
		payload := json.RawMessage(r.Payload)
		if len(payload) == 0 {
			payload = json.RawMessage(`{}`)
		}
		docs = append(docs, RagCorpusDocument{
			DocumentID: r.ID, TypeID: r.TypeID, TypeCode: r.TypeCode, TypeName: r.TypeName,
			Number: r.Number, Title: r.Title, Status: r.Status, AuthorSub: r.AuthorSub,
			Payload: payload, Content: content, SearchText: searchText,
			ReaderRoles: r.ReaderRoles, WriterRoles: r.WriterRoles,
		})
	}

	users, err := authUsers.ListTenantUsers(ctx, tenant)
	if err != nil {
		return nil, err
	}
	access := buildUserAccess(tenant, users, rows)
	return &RagCorpusResponse{
		Tenant: tenant, FetchedAt: time.Now().UTC(),
		Documents: docs, Users: access,
	}, nil
}

func extractSearchText(rag json.RawMessage) string {
	var m map[string]any
	if err := json.Unmarshal(rag, &m); err != nil {
		return ""
	}
	if s, ok := m["search_text"].(string); ok {
		return s
	}
	return ""
}

func buildUserAccess(tenant string, users []clients.AuthTenantUser, rows []repositories.RagCorpusRow) []RagCorpusUserAccess {
	out := make([]RagCorpusUserAccess, 0, len(users))
	for _, u := range users {
		login := u.Username
		if login != "" && !containsAt(login) {
			login = u.Username + "@" + tenant
		}
		claims := &models.Claims{Sub: u.KeycloakID, RealmRoles: u.Roles}
		visible := make([]string, 0, len(rows))
		for _, r := range rows {
			if docVisibleToUser(claims, r) {
				visible = append(visible, r.ID)
			}
		}
		out = append(out, RagCorpusUserAccess{
			Login: login, Username: u.Username, KeycloakID: u.KeycloakID, Roles: u.Roles,
			VisibleDocumentIDs: visible, ExpectedCount: len(visible),
		})
	}
	return out
}

func docVisibleToUser(c *models.Claims, r repositories.RagCorpusRow) bool {
	if c == nil {
		return false
	}
	if models.CanAdminSED(c) {
		return true
	}
	if r.AuthorSub != "" && r.AuthorSub == c.Sub {
		return true
	}
	return models.RolesOverlap(c.RealmRoles, r.ReaderRoles)
}

func containsAt(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '@' {
			return true
		}
	}
	return false
}
