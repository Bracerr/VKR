package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/industrial-sed/procurement-service/internal/clients"
	"github.com/industrial-sed/procurement-service/internal/models"
)

func histPayload(m map[string]any) json.RawMessage {
	b, _ := json.Marshal(m)
	return b
}

// --- Suppliers ---

func (a *App) ListSuppliers(ctx context.Context, tenant string) ([]models.Supplier, error) {
	return a.Store.ListSuppliers(ctx, nil, tenant)
}

func (a *App) CreateSupplier(ctx context.Context, tenant, actorSub, code, name string, inn, kpp *string, contacts json.RawMessage, active bool) (*models.Supplier, error) {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" {
		return nil, ErrValidation
	}
	if len(contacts) == 0 {
		contacts = json.RawMessage(`{}`)
	}
	sp := &models.Supplier{
		ID:         uuid.New(),
		TenantCode: tenant,
		Code:       code,
		Name:       name,
		INN:        inn,
		KPP:        kpp,
		Contacts:   contacts,
		Active:     active,
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.CreateSupplier(ctx, tx, sp); err != nil {
		return nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "supplier", sp.ID, actorSub, "CREATE", histPayload(map[string]any{"code": code}))
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return sp, nil
}

func (a *App) UpdateSupplier(ctx context.Context, tenant, actorSub string, id uuid.UUID, code, name string, inn, kpp *string, contacts json.RawMessage, active bool) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	sp, err := a.Store.GetSupplier(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if sp == nil {
		return ErrNotFound
	}
	if len(contacts) == 0 {
		contacts = json.RawMessage(`{}`)
	}
	sp.Code, sp.Name, sp.INN, sp.KPP, sp.Contacts, sp.Active = code, name, inn, kpp, contacts, active
	if err := a.Store.UpdateSupplier(ctx, tx, sp); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "supplier", id, actorSub, "UPDATE", nil)
	return tx.Commit(ctx)
}

func (a *App) DeleteSupplier(ctx context.Context, tenant, actorSub string, id uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.DeleteSupplier(ctx, tx, tenant, id); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "supplier", id, actorSub, "DELETE", nil)
	return tx.Commit(ctx)
}

// --- Purchase Requests ---

func (a *App) ListPR(ctx context.Context, tenant string, status *string) ([]models.PurchaseRequest, error) {
	return a.Store.ListPR(ctx, nil, tenant, status)
}

func (a *App) GetPRDetail(ctx context.Context, tenant string, id uuid.UUID) (*models.PurchaseRequest, []models.PurchaseRequestLine, error) {
	pr, err := a.Store.GetPR(ctx, nil, tenant, id)
	if err != nil {
		return nil, nil, err
	}
	if pr == nil {
		return nil, nil, ErrNotFound
	}
	lines, err := a.Store.ListPRLines(ctx, nil, id)
	if err != nil {
		return nil, nil, err
	}
	return pr, lines, nil
}

func (a *App) CreatePR(ctx context.Context, tenant, actorSub string, neededBy *time.Time, note *string) (*models.PurchaseRequest, error) {
	id := uuid.New()
	num := fmt.Sprintf("PR-%s", strings.ReplaceAll(id.String(), "-", "")[:12])
	pr := &models.PurchaseRequest{
		ID:            id,
		TenantCode:    tenant,
		Number:        num,
		Status:        "DRAFT",
		CreatedBySub:  actorSub,
		NeededBy:      neededBy,
		Note:          note,
		SedDocumentID: nil,
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.CreatePR(ctx, tx, pr); err != nil {
		return nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "pr", pr.ID, actorSub, "CREATE", histPayload(map[string]any{"number": num}))
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return pr, nil
}

func (a *App) AddPRLine(ctx context.Context, tenant, actorSub string, prID uuid.UUID, lineNo int, productID uuid.UUID, qty, uom string, whID, binID *uuid.UUID, note *string) (*models.PurchaseRequestLine, error) {
	q, err := decimal.NewFromString(qty)
	if err != nil || q.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: qty", ErrValidation)
	}
	if strings.TrimSpace(uom) == "" {
		uom = "pcs"
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	pr, err := a.Store.LockPR(ctx, tx, tenant, prID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, ErrNotFound
	}
	if pr.Status != "DRAFT" {
		return nil, ErrWrongState
	}
	ln := &models.PurchaseRequestLine{
		ID:                uuid.New(),
		PRID:              prID,
		LineNo:            lineNo,
		ProductID:         productID,
		Qty:               q.StringFixed(8),
		UOM:               uom,
		TargetWarehouseID: whID,
		TargetBinID:       binID,
		Note:              note,
	}
	if err := a.Store.AddPRLine(ctx, tx, ln); err != nil {
		return nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "pr", prID, actorSub, "ADD_LINE", histPayload(map[string]any{"line_no": lineNo}))
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ln, nil
}

func (a *App) SubmitPR(ctx context.Context, tenant, actorSub, bearer string, prID uuid.UUID, sedDocTypeID uuid.UUID, title string) error {
	if bearer == "" {
		return ErrValidation
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	pr, err := a.Store.LockPR(ctx, tx, tenant, prID)
	if err != nil {
		return err
	}
	if pr == nil {
		return ErrNotFound
	}
	if pr.Status != "DRAFT" {
		return ErrWrongState
	}
	lines, err := a.Store.ListPRLines(ctx, tx, prID)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return ErrValidation
	}
	payload, _ := json.Marshal(map[string]any{"kind": "purchase_request", "pr_id": prID.String()})
	doc, err := a.SED.CreateDocument(ctx, bearer, sedDocTypeID, title, payload)
	if err != nil {
		return err
	}
	if err := a.SED.SubmitDocument(ctx, bearer, doc.ID); err != nil {
		return err
	}
	sedID := doc.ID
	if err := a.Store.UpdatePRStatusSed(ctx, tx, tenant, prID, "SUBMITTED", &sedID); err != nil {
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "pr", prID, actorSub, "SUBMIT_SED", histPayload(map[string]any{"sed_document_id": sedID.String()}))
	return tx.Commit(ctx)
}

func (a *App) CancelPR(ctx context.Context, tenant, actorSub string, prID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	pr, err := a.Store.LockPR(ctx, tx, tenant, prID)
	if err != nil {
		return err
	}
	if pr == nil {
		return ErrNotFound
	}
	switch pr.Status {
	case "DRAFT", "SUBMITTED", "APPROVED":
	default:
		return ErrWrongState
	}
	if err := a.Store.UpdatePRStatusSed(ctx, tx, tenant, prID, "CANCELLED", pr.SedDocumentID); err != nil {
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "pr", prID, actorSub, "CANCEL", nil)
	return tx.Commit(ctx)
}

// --- Purchase Orders ---

func (a *App) ListPO(ctx context.Context, tenant string, status *string) ([]models.PurchaseOrder, error) {
	return a.Store.ListPO(ctx, nil, tenant, status)
}

func (a *App) GetPODetail(ctx context.Context, tenant string, id uuid.UUID) (*models.PurchaseOrder, []models.PurchaseOrderLine, error) {
	po, err := a.Store.GetPO(ctx, nil, tenant, id)
	if err != nil {
		return nil, nil, err
	}
	if po == nil {
		return nil, nil, ErrNotFound
	}
	lines, err := a.Store.ListPOLines(ctx, nil, id)
	if err != nil {
		return nil, nil, err
	}
	return po, lines, nil
}

func (a *App) CreatePO(ctx context.Context, tenant, actorSub string, supplierID uuid.UUID, currency string, expectedAt *time.Time, sourcePRID *uuid.UUID) (*models.PurchaseOrder, error) {
	if currency == "" {
		currency = "RUB"
	}
	id := uuid.New()
	num := fmt.Sprintf("PO-%s", strings.ReplaceAll(id.String(), "-", "")[:12])
	po := &models.PurchaseOrder{
		ID:            id,
		TenantCode:    tenant,
		Number:        num,
		SupplierID:    supplierID,
		Status:        "DRAFT",
		CreatedBySub:  actorSub,
		Currency:      currency,
		ExpectedAt:    expectedAt,
		SedDocumentID: nil,
		SourcePRID:    sourcePRID,
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	sp, err := a.Store.GetSupplier(ctx, tx, tenant, supplierID)
	if err != nil {
		return nil, err
	}
	if sp == nil {
		return nil, ErrNotFound
	}
	if err := a.Store.CreatePO(ctx, tx, po); err != nil {
		return nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "po", po.ID, actorSub, "CREATE", histPayload(map[string]any{"number": num}))
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return po, nil
}

func (a *App) CreatePOFromPR(ctx context.Context, tenant, actorSub string, prID uuid.UUID, supplierID uuid.UUID) (*models.PurchaseOrder, error) {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	pr, err := a.Store.GetPR(ctx, tx, tenant, prID)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, ErrNotFound
	}
	lines, err := a.Store.ListPRLines(ctx, tx, prID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrValidation
	}
	src := pr.ID
	po, err := a.CreatePO(ctx, tenant, actorSub, supplierID, "RUB", nil, &src)
	if err != nil {
		return nil, err
	}
	// отдельной транзакцией добавляем lines в PO
	tx2, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx2.Rollback(ctx) }()
	for _, ln := range lines {
		pl := &models.PurchaseOrderLine{
			ID:                uuid.New(),
			POID:              po.ID,
			LineNo:            ln.LineNo,
			ProductID:         ln.ProductID,
			QtyOrdered:        ln.Qty,
			QtyReceived:       "0",
			Price:             "0",
			VATRate:           "0",
			TargetWarehouseID: ln.TargetWarehouseID,
			TargetBinID:       ln.TargetBinID,
		}
		if err := a.Store.AddPOLine(ctx, tx2, pl); err != nil {
			return nil, err
		}
	}
	_ = a.Store.InsertHistory(ctx, tx2, tenant, "po", po.ID, actorSub, "FROM_PR", histPayload(map[string]any{"pr_id": prID.String()}))
	if err := tx2.Commit(ctx); err != nil {
		return nil, err
	}
	return po, nil
}

func (a *App) AddPOLine(ctx context.Context, tenant, actorSub string, poID uuid.UUID, lineNo int, productID uuid.UUID, qtyOrdered, price, vatRate string, whID, binID *uuid.UUID) (*models.PurchaseOrderLine, error) {
	q, err := decimal.NewFromString(qtyOrdered)
	if err != nil || q.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("%w: qty_ordered", ErrValidation)
	}
	if strings.TrimSpace(price) == "" {
		price = "0"
	}
	if strings.TrimSpace(vatRate) == "" {
		vatRate = "0"
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	po, err := a.Store.LockPO(ctx, tx, tenant, poID)
	if err != nil {
		return nil, err
	}
	if po == nil {
		return nil, ErrNotFound
	}
	if po.Status != "DRAFT" {
		return nil, ErrWrongState
	}
	ln := &models.PurchaseOrderLine{
		ID:                uuid.New(),
		POID:              poID,
		LineNo:            lineNo,
		ProductID:         productID,
		QtyOrdered:        q.StringFixed(8),
		QtyReceived:       "0",
		Price:             price,
		VATRate:           vatRate,
		TargetWarehouseID: whID,
		TargetBinID:       binID,
	}
	if err := a.Store.AddPOLine(ctx, tx, ln); err != nil {
		return nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "po", poID, actorSub, "ADD_LINE", histPayload(map[string]any{"line_no": lineNo}))
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ln, nil
}

func (a *App) SubmitPO(ctx context.Context, tenant, actorSub, bearer string, poID uuid.UUID, sedDocTypeID uuid.UUID, title string) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	po, err := a.Store.LockPO(ctx, tx, tenant, poID)
	if err != nil {
		return err
	}
	if po == nil {
		return ErrNotFound
	}
	if po.Status != "DRAFT" {
		return ErrWrongState
	}
	lines, err := a.Store.ListPOLines(ctx, tx, poID)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return ErrValidation
	}
	payload, _ := json.Marshal(map[string]any{"kind": "purchase_order", "po_id": poID.String()})
	doc, err := a.SED.CreateDocument(ctx, bearer, sedDocTypeID, title, payload)
	if err != nil {
		return err
	}
	if err := a.SED.SubmitDocument(ctx, bearer, doc.ID); err != nil {
		return err
	}
	sedID := doc.ID
	if err := a.Store.UpdatePOStatusSed(ctx, tx, tenant, poID, "SUBMITTED", &sedID); err != nil {
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "po", poID, actorSub, "SUBMIT_SED", histPayload(map[string]any{"sed_document_id": sedID.String()}))
	return tx.Commit(ctx)
}

func (a *App) ReleasePO(ctx context.Context, tenant, actorSub string, poID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	po, err := a.Store.LockPO(ctx, tx, tenant, poID)
	if err != nil {
		return err
	}
	if po == nil {
		return ErrNotFound
	}
	if po.Status != "APPROVED" {
		return ErrWrongState
	}
	if err := a.Store.UpdatePOStatusSed(ctx, tx, tenant, poID, "RELEASED", po.SedDocumentID); err != nil {
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "po", poID, actorSub, "RELEASE", nil)
	return tx.Commit(ctx)
}

func (a *App) CancelPO(ctx context.Context, tenant, actorSub string, poID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	po, err := a.Store.LockPO(ctx, tx, tenant, poID)
	if err != nil {
		return err
	}
	if po == nil {
		return ErrNotFound
	}
	switch po.Status {
	case "DRAFT", "SUBMITTED", "APPROVED", "RELEASED", "PARTIALLY_RECEIVED":
	default:
		return ErrWrongState
	}
	if err := a.Store.UpdatePOStatusSed(ctx, tx, tenant, poID, "CANCELLED", po.SedDocumentID); err != nil {
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "po", poID, actorSub, "CANCEL", nil)
	return tx.Commit(ctx)
}

// ReceivePO выполняет приемку по PO в warehouse.
func (a *App) ReceivePO(ctx context.Context, tenant, actorSub string, poID uuid.UUID, warehouseID uuid.UUID, binID uuid.UUID) (uuid.UUID, error) {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	po, err := a.Store.LockPO(ctx, tx, tenant, poID)
	if err != nil {
		return uuid.Nil, err
	}
	if po == nil {
		return uuid.Nil, ErrNotFound
	}
	if po.Status != "RELEASED" && po.Status != "PARTIALLY_RECEIVED" {
		return uuid.Nil, ErrWrongState
	}
	lines, err := a.Store.ListPOLines(ctx, tx, poID)
	if err != nil {
		return uuid.Nil, err
	}
	if len(lines) == 0 {
		return uuid.Nil, ErrValidation
	}
	req := &clients.ReceiptRequest{
		WarehouseID: warehouseID.String(),
		BinID:       binID.String(),
	}
	for _, ln := range lines {
		need := ln.QtyOrdered
		req.Lines = append(req.Lines, clients.ReceiptLine{ProductID: ln.ProductID.String(), Qty: need})
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	idem := "po-receipt-" + poID.String()
	whDocID, err := a.WH.Receipt(ctx, tenant, req, idem)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrWarehouse, err)
	}
	tx2, err := a.Store.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx2.Rollback(ctx) }()
	r := &models.Receipt{ID: uuid.New(), TenantCode: tenant, POID: poID, WarehouseDocumentID: whDocID, Status: "POSTED", PostedAt: time.Now().UTC()}
	if err := a.Store.InsertReceipt(ctx, tx2, r); err != nil {
		return uuid.Nil, err
	}
	// в MVP считаем что всё принято
	if err := a.Store.UpdatePOStatusSed(ctx, tx2, tenant, poID, "RECEIVED", po.SedDocumentID); err != nil {
		return uuid.Nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx2, tenant, "po", poID, actorSub, "RECEIVE", histPayload(map[string]any{"warehouse_document_id": whDocID.String()}))
	if err := tx2.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return whDocID, nil
}

// HandleSedDocumentSigned callback от SED после подписи закупочного документа.
func (a *App) HandleSedDocumentSigned(ctx context.Context, tenant string, documentID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	pr, err := a.Store.FindPRBySedDocument(ctx, tx, tenant, documentID)
	if err != nil {
		return err
	}
	if pr != nil {
		if pr.Status == "APPROVED" {
			return nil
		}
		if pr.Status != "SUBMITTED" {
			return ErrWrongState
		}
		if err := a.Store.UpdatePRStatusSed(ctx, tx, tenant, pr.ID, "APPROVED", pr.SedDocumentID); err != nil {
			return err
		}
		_ = a.Store.InsertHistory(ctx, tx, tenant, "pr", pr.ID, "sed_callback", "APPROVED_SED", histPayload(map[string]any{"document_id": documentID.String()}))
		return tx.Commit(ctx)
	}
	po, err := a.Store.FindPOBySedDocument(ctx, tx, tenant, documentID)
	if err != nil {
		return err
	}
	if po != nil {
		if po.Status == "APPROVED" {
			return nil
		}
		if po.Status != "SUBMITTED" {
			return ErrWrongState
		}
		if err := a.Store.UpdatePOStatusSed(ctx, tx, tenant, po.ID, "APPROVED", po.SedDocumentID); err != nil {
			return err
		}
		_ = a.Store.InsertHistory(ctx, tx, tenant, "po", po.ID, "sed_callback", "APPROVED_SED", histPayload(map[string]any{"document_id": documentID.String()}))
		return tx.Commit(ctx)
	}
	return ErrNotFound
}

