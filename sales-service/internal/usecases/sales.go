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

	"github.com/industrial-sed/sales-service/internal/clients"
	"github.com/industrial-sed/sales-service/internal/models"
)

func histPayload(m map[string]any) json.RawMessage {
	b, _ := json.Marshal(m)
	return b
}

func emptyArrJSON() json.RawMessage { return json.RawMessage(`[]`) }

// --- Customers ---

func (a *App) ListCustomers(ctx context.Context, tenant string) ([]models.Customer, error) {
	return a.Store.ListCustomers(ctx, nil, tenant)
}

func (a *App) CreateCustomer(ctx context.Context, tenant, actorSub, code, name string, contacts json.RawMessage, active bool) (*models.Customer, error) {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" {
		return nil, ErrValidation
	}
	if len(contacts) == 0 {
		contacts = json.RawMessage(`{}`)
	}
	c := &models.Customer{ID: uuid.New(), TenantCode: tenant, Code: code, Name: name, Contacts: contacts, Active: active}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.CreateCustomer(ctx, tx, c); err != nil {
		return nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "customer", c.ID, actorSub, "CREATE", histPayload(map[string]any{"code": code}))
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

func (a *App) UpdateCustomer(ctx context.Context, tenant, actorSub string, id uuid.UUID, code, name string, contacts json.RawMessage, active bool) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	c, err := a.Store.GetCustomer(ctx, tx, tenant, id)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrNotFound
	}
	if len(contacts) == 0 {
		contacts = json.RawMessage(`{}`)
	}
	c.Code, c.Name, c.Contacts, c.Active = code, name, contacts, active
	if err := a.Store.UpdateCustomer(ctx, tx, c); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "customer", id, actorSub, "UPDATE", nil)
	return tx.Commit(ctx)
}

func (a *App) DeleteCustomer(ctx context.Context, tenant, actorSub string, id uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := a.Store.DeleteCustomer(ctx, tx, tenant, id); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "customer", id, actorSub, "DELETE", nil)
	return tx.Commit(ctx)
}

// --- Sales Orders ---

func (a *App) ListSO(ctx context.Context, tenant string, status *string) ([]models.SalesOrder, error) {
	return a.Store.ListSO(ctx, nil, tenant, status)
}

func (a *App) GetSODetail(ctx context.Context, tenant string, id uuid.UUID) (*models.SalesOrder, []models.SalesOrderLine, error) {
	so, err := a.Store.GetSO(ctx, nil, tenant, id)
	if err != nil {
		return nil, nil, err
	}
	if so == nil {
		return nil, nil, ErrNotFound
	}
	lines, err := a.Store.ListSOLines(ctx, nil, id)
	if err != nil {
		return nil, nil, err
	}
	return so, lines, nil
}

func (a *App) CreateSO(ctx context.Context, tenant, actorSub string, customerID uuid.UUID, whID, binID *uuid.UUID, note *string) (*models.SalesOrder, error) {
	id := uuid.New()
	num := fmt.Sprintf("SO-%s", strings.ReplaceAll(id.String(), "-", "")[:12])
	so := &models.SalesOrder{
		ID: id, TenantCode: tenant, Number: num, Status: "DRAFT",
		CustomerID: customerID, CreatedBySub: actorSub,
		ShipFromWarehouseID: whID, ShipFromBinID: binID,
		Note: note, SedDocumentID: nil, Reservations: emptyArrJSON(),
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	c, err := a.Store.GetCustomer(ctx, tx, tenant, customerID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, ErrNotFound
	}
	if err := a.Store.CreateSO(ctx, tx, so); err != nil {
		return nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "so", so.ID, actorSub, "CREATE", histPayload(map[string]any{"number": num}))
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return so, nil
}

func (a *App) AddSOLine(ctx context.Context, tenant, actorSub string, soID uuid.UUID, lineNo int, productID uuid.UUID, qty, uom string, note *string) (*models.SalesOrderLine, error) {
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
	so, err := a.Store.LockSO(ctx, tx, tenant, soID)
	if err != nil {
		return nil, err
	}
	if so == nil {
		return nil, ErrNotFound
	}
	if so.Status != "DRAFT" {
		return nil, ErrWrongState
	}
	ln := &models.SalesOrderLine{
		ID: uuid.New(), SOID: soID, LineNo: lineNo, ProductID: productID,
		Qty: q.StringFixed(8), UOM: uom,
		ReservedQty: "0", ShippedQty: "0", Note: note,
	}
	if err := a.Store.AddSOLine(ctx, tx, ln); err != nil {
		return nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "so", soID, actorSub, "ADD_LINE", histPayload(map[string]any{"line_no": lineNo}))
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ln, nil
}

func (a *App) SubmitSO(ctx context.Context, tenant, actorSub, bearer string, soID uuid.UUID, sedDocTypeID uuid.UUID, title string) error {
	if bearer == "" {
		return ErrValidation
	}
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	so, err := a.Store.LockSO(ctx, tx, tenant, soID)
	if err != nil {
		return err
	}
	if so == nil {
		return ErrNotFound
	}
	if so.Status != "DRAFT" {
		return ErrWrongState
	}
	lines, err := a.Store.ListSOLines(ctx, tx, soID)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return ErrValidation
	}
	payload, _ := json.Marshal(map[string]any{"kind": "sales_order", "so_id": soID.String()})
	doc, err := a.SED.CreateDocument(ctx, bearer, sedDocTypeID, title, payload)
	if err != nil {
		return err
	}
	if err := a.SED.SubmitDocument(ctx, bearer, doc.ID); err != nil {
		return err
	}
	sedID := doc.ID
	if err := a.Store.UpdateSOStatusSedAndReservations(ctx, tx, tenant, soID, "SUBMITTED", &sedID, so.Reservations); err != nil {
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "so", soID, actorSub, "SUBMIT_SED", histPayload(map[string]any{"sed_document_id": sedID.String()}))
	return tx.Commit(ctx)
}

func (a *App) ReleaseSO(ctx context.Context, tenant, actorSub string, soID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	so, err := a.Store.LockSO(ctx, tx, tenant, soID)
	if err != nil {
		return err
	}
	if so == nil {
		return ErrNotFound
	}
	if so.Status != "APPROVED" {
		return ErrWrongState
	}
	if err := a.Store.UpdateSOStatusSedAndReservations(ctx, tx, tenant, soID, "RELEASED", so.SedDocumentID, so.Reservations); err != nil {
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "so", soID, actorSub, "RELEASE", nil)
	return tx.Commit(ctx)
}

func (a *App) CancelSO(ctx context.Context, tenant, actorSub string, soID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	so, err := a.Store.LockSO(ctx, tx, tenant, soID)
	if err != nil {
		return err
	}
	if so == nil {
		return ErrNotFound
	}
	switch so.Status {
	case "DRAFT", "SUBMITTED", "APPROVED", "RELEASED":
	default:
		return ErrWrongState
	}
	if err := a.Store.UpdateSOStatusSedAndReservations(ctx, tx, tenant, soID, "CANCELLED", so.SedDocumentID, so.Reservations); err != nil {
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "so", soID, actorSub, "CANCEL", nil)
	return tx.Commit(ctx)
}

// ReserveSO создаёт резервы в warehouse-service и сохраняет reservation_ids.
func (a *App) ReserveSO(ctx context.Context, tenant, actorSub string, soID uuid.UUID) ([]uuid.UUID, error) {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	so, err := a.Store.LockSO(ctx, tx, tenant, soID)
	if err != nil {
		return nil, err
	}
	if so == nil {
		return nil, ErrNotFound
	}
	if so.Status != "RELEASED" {
		return nil, ErrWrongState
	}
	if so.ShipFromWarehouseID == nil || so.ShipFromBinID == nil {
		return nil, fmt.Errorf("%w: ship_from_warehouse_id/bin_id", ErrValidation)
	}
	// идемпотентность: если уже есть reservations
	var existing []string
	_ = json.Unmarshal(so.Reservations, &existing)
	if len(existing) > 0 {
		var ids []uuid.UUID
		for _, s := range existing {
			id, _ := uuid.Parse(s)
			if id != uuid.Nil {
				ids = append(ids, id)
			}
		}
		return ids, tx.Commit(ctx)
	}
	lines, err := a.Store.ListSOLines(ctx, tx, soID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrValidation
	}
	// коммитим, чтобы не держать tx во время HTTP
	snapWh := so.ShipFromWarehouseID.String()
	snapBin := so.ShipFromBinID.String()
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	var ids []uuid.UUID
	for _, ln := range lines {
		req := &clients.ReservationRequest{
			WarehouseID: snapWh,
			BinID:       snapBin,
			ProductID:   ln.ProductID.String(),
			Qty:         ln.Qty,
			Reason:      "SALES_RESERVE",
			DocRef:      so.Number,
		}
		idem := "so-reserve-" + soID.String() + "-" + fmt.Sprintf("%d", ln.LineNo)
		rid, err := a.WH.CreateReservation(ctx, tenant, req, idem)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrWarehouse, err)
		}
		ids = append(ids, rid)
	}
	// записываем reservations
	tx2, err := a.Store.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx2.Rollback(ctx) }()
	so2, err := a.Store.LockSO(ctx, tx2, tenant, soID)
	if err != nil {
		return nil, err
	}
	if so2 == nil {
		return nil, ErrNotFound
	}
	raw, _ := json.Marshal(func() []string {
		out := make([]string, 0, len(ids))
		for _, id := range ids {
			out = append(out, id.String())
		}
		return out
	}())
	if err := a.Store.UpdateSOStatusSedAndReservations(ctx, tx2, tenant, soID, so2.Status, so2.SedDocumentID, raw); err != nil {
		return nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx2, tenant, "so", soID, actorSub, "RESERVE", histPayload(map[string]any{"n": len(ids)}))
	if err := tx2.Commit(ctx); err != nil {
		return nil, err
	}
	return ids, nil
}

// ShipSO: release reservations (to avoid reserved_qty>quantity constraint) and issue.
func (a *App) ShipSO(ctx context.Context, tenant, actorSub string, soID uuid.UUID) (uuid.UUID, error) {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	so, err := a.Store.LockSO(ctx, tx, tenant, soID)
	if err != nil {
		return uuid.Nil, err
	}
	if so == nil {
		return uuid.Nil, ErrNotFound
	}
	if so.Status != "RELEASED" {
		return uuid.Nil, ErrWrongState
	}
	sh, err := a.Store.GetShipmentBySO(ctx, tx, tenant, soID)
	if err != nil {
		return uuid.Nil, err
	}
	if sh != nil && sh.WarehouseDocumentID != nil {
		return *sh.WarehouseDocumentID, tx.Commit(ctx)
	}
	if so.ShipFromWarehouseID == nil || so.ShipFromBinID == nil {
		return uuid.Nil, fmt.Errorf("%w: ship_from_warehouse_id/bin_id", ErrValidation)
	}
	var resIDs []string
	_ = json.Unmarshal(so.Reservations, &resIDs)
	if len(resIDs) == 0 {
		return uuid.Nil, fmt.Errorf("%w: нет резервов", ErrValidation)
	}
	number := so.Number
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	var rids []uuid.UUID
	for _, s := range resIDs {
		id, err := uuid.Parse(s)
		if err == nil && id != uuid.Nil {
			rids = append(rids, id)
		}
	}
	// ship = issue-from-reservations (атомарно уменьшает quantity и reserved_qty)
	docID, err := a.WH.IssueFromReservations(ctx, tenant, rids, "so-ship-"+soID.String())
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrWarehouse, err)
	}
	// persist shipment + status
	tx2, err := a.Store.BeginTx(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer func() { _ = tx2.Rollback(ctx) }()
	so2, err := a.Store.LockSO(ctx, tx2, tenant, soID)
	if err != nil {
		return uuid.Nil, err
	}
	if so2 == nil {
		return uuid.Nil, ErrNotFound
	}
	d := docID
	shp := &models.Shipment{ID: uuid.New(), TenantCode: tenant, SOID: soID, WarehouseDocumentID: &d, Status: "POSTED", PostedAt: time.Now().UTC()}
	if err := a.Store.InsertShipment(ctx, tx2, shp); err != nil {
		return uuid.Nil, err
	}
	if err := a.Store.UpdateSOStatusSedAndReservations(ctx, tx2, tenant, soID, "SHIPPED", so2.SedDocumentID, so2.Reservations); err != nil {
		return uuid.Nil, err
	}
	_ = a.Store.InsertHistory(ctx, tx2, tenant, "so", soID, actorSub, "SHIP", histPayload(map[string]any{"warehouse_document_id": docID.String(), "so_number": number}))
	if err := tx2.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	if a.Trace != nil {
		bg, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = a.Trace.LinkEntityToWarehouseDoc(bg, tenant, "SO", soID.String(), number, docID.String(), "so-link-"+soID.String())
	}
	return docID, nil
}

// HandleSedDocumentSigned callback от SED после подписи документа по SO.
func (a *App) HandleSedDocumentSigned(ctx context.Context, tenant string, documentID uuid.UUID) error {
	tx, err := a.Store.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	so, err := a.Store.FindSOBySedDocument(ctx, tx, tenant, documentID)
	if err != nil {
		return err
	}
	if so == nil {
		return ErrNotFound
	}
	if so.Status == "APPROVED" {
		return tx.Commit(ctx)
	}
	if so.Status != "SUBMITTED" {
		return ErrWrongState
	}
	if err := a.Store.UpdateSOStatusSedAndReservations(ctx, tx, tenant, so.ID, "APPROVED", so.SedDocumentID, so.Reservations); err != nil {
		return err
	}
	_ = a.Store.InsertHistory(ctx, tx, tenant, "so", so.ID, "sed_callback", "APPROVED_SED", histPayload(map[string]any{"document_id": documentID.String()}))
	return tx.Commit(ctx)
}

