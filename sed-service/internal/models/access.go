package models

import "strings"

// Роли чтения/записи по типам документов (realm).
const (
	RoleDocReadProcurement  = "doc_read_procurement"
	RoleDocReadSales        = "doc_read_sales"
	RoleDocReadProduction   = "doc_read_production"
	RoleDocReadWarehouse    = "doc_read_warehouse"
	RoleDocReadFinance      = "doc_read_finance"
	RoleDocWriteProcurement = "doc_write_procurement"
	RoleDocWriteSales       = "doc_write_sales"
	RoleDocWriteProduction  = "doc_write_production"
)

var allDocReadRoles = []string{
	RoleDocReadProcurement, RoleDocReadSales, RoleDocReadProduction,
	RoleDocReadWarehouse, RoleDocReadFinance,
}

// HasAnyDocReadRole есть ли у пользователя хотя бы одна doc_read_* роль.
func HasAnyDocReadRole(c *Claims) bool {
	if c == nil {
		return false
	}
	for _, r := range c.RealmRoles {
		for _, dr := range allDocReadRoles {
			if r == dr {
				return true
			}
		}
	}
	return false
}

// RolesOverlap пересечение ролей пользователя с allowed.
func RolesOverlap(userRoles, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}
	for _, u := range userRoles {
		for _, a := range allowed {
			if u == a {
				return true
			}
		}
	}
	return false
}

// DefaultACLForDocumentType подставляет reader/writer при пустых массивах.
func DefaultACLForDocumentType(code, warehouseAction string) (readers, writers []string) {
	c := strings.ToUpper(code)
	switch {
	case strings.HasPrefix(c, "PURCHASE_") || strings.HasPrefix(c, "SUPPLIER_"):
		return []string{RoleDocReadProcurement, RoleDocReadFinance}, []string{RoleDocWriteProcurement}
	case strings.HasPrefix(c, "SALES_") || strings.HasPrefix(c, "SHIPMENT_"):
		return []string{RoleDocReadSales}, []string{RoleDocWriteSales}
	case strings.HasPrefix(c, "BOM_") || strings.HasPrefix(c, "ROUTING_"):
		return []string{RoleDocReadProduction}, []string{RoleDocWriteProduction}
	default:
		if warehouseAction != "" && warehouseAction != "NONE" {
			return []string{RoleDocReadWarehouse}, []string{RoleSedAuthor}
		}
	}
	return nil, nil
}

// CanReadDocument доступ на чтение (author всегда видит свои; pendingApprover — отдельно).
func CanReadDocument(c *Claims, d *Document, dt *DocumentType, pendingApprover bool) bool {
	if c == nil || d == nil || dt == nil {
		return false
	}
	if CanAdminSED(c) {
		return true
	}
	if d.AuthorSub == c.Sub {
		return true
	}
	if pendingApprover {
		return true
	}
	return RolesOverlap(c.RealmRoles, dt.ReaderRoles)
}

// CanCreateDocument создание документа данного типа.
func CanCreateDocument(c *Claims, dt *DocumentType) bool {
	if c == nil || dt == nil {
		return false
	}
	if CanAdminSED(c) {
		return true
	}
	if !CanAuthor(c) {
		return false
	}
	return RolesOverlap(c.RealmRoles, dt.WriterRoles)
}

// CanWriteOwnDraft правка своего черновика.
func CanWriteOwnDraft(c *Claims, d *Document) bool {
	if c == nil || d == nil {
		return false
	}
	if CanAdminSED(c) {
		return true
	}
	return d.AuthorSub == c.Sub && CanAuthor(c)
}
