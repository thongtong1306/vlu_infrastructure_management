package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"vlu_infrastructure_management/models"
)

// Convenience: default ctx if you don't pass one.
func ctxOrBackground(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}

// =========================
// Equipment (log_lab_equipment_master)
// =========================

//func EquipGetByID(ctx context.Context, id uint64) (*models.Equipment, error) {
//	var e models.Equipment
//	err := srv.DB.GetContext(ctxOrBackground(ctx), &e, `
//		SELECT id, sku, name, serial_no, location, category, status, purchase_at, vendor, warranty_exp, created_at, updated_at
//		FROM log_lab_equipment_master
//		WHERE id=? LIMIT 1`, id)
//	if err == sql.ErrNoRows {
//		return nil, nil
//	}
//	return &e, err
//}
//
//func EquipGetBySKU(ctx context.Context, sku string) (*models.Equipment, error) {
//	var e models.Equipment
//	err := srv.DB.GetContext(ctxOrBackground(ctx), &e, `
//		SELECT id, sku, name, serial_no, location, category, status, purchase_at, vendor, warranty_exp, created_at, updated_at
//		FROM log_lab_equipment_master
//		WHERE sku=? LIMIT 1`, sku)
//	if err == sql.ErrNoRows {
//		return nil, nil
//	}
//	return &e, err
//}
//
//func EquipList(ctx context.Context, q, status string, limit, offset int) ([]models.Equipment, error) {
//	if limit <= 0 {
//		limit = 50
//	}
//	args := []interface{}{}
//	sqlStr := `
//		SELECT id, sku, name, serial_no, location, category, status, purchase_at, vendor, warranty_exp, created_at, updated_at
//		FROM log_lab_equipment_master
//		WHERE 1=1`
//	if q != "" {
//		sqlStr += ` AND (sku LIKE ? OR name LIKE ?)`
//		args = append(args, "%"+q+"%", "%"+q+"%")
//	}
//	if status != "" {
//		sqlStr += ` AND status=?`
//		args = append(args, status)
//	}
//	sqlStr += ` ORDER BY updated_at DESC LIMIT ? OFFSET ?`
//	args = append(args, limit, offset)
//
//	var list []models.Equipment
//	return list, srv.DB.SelectContext(ctxOrBackground(ctx), &list, sqlStr, args...)
//}
//
//func EquipCreate(ctx context.Context, e *models.Equipment) (uint64, error) {
//	res, err := srv.DB.ExecContext(ctxOrBackground(ctx), `
//		INSERT INTO log_lab_equipment_master
//			(sku, name, serial_no, location, category, status, purchase_at, vendor, warranty_exp)
//		VALUES (?,?,?,?,?,?,?,?,?)`,
//		e.SKU, e.Name, e.SerialNo, e.Location, e.Category, e.Status, e.PurchaseAt, e.Vendor, e.WarrantyExp,
//	)
//	if err != nil {
//		return 0, err
//	}
//	id, _ := res.LastInsertId()
//	return uint64(id), nil
//}
//
//func EquipSetStatus(ctx context.Context, id uint64, status string) error {
//	_, err := srv.DB.ExecContext(ctxOrBackground(ctx),
//		`UPDATE log_lab_equipment_master SET status=? WHERE id=?`, status, id)
//	return err
//}

// =========================
// Borrowing (log_lab_borrow_records)
// =========================

func BorrowCreate(ctx context.Context, itemID uint64, borrowerName string, borrowerID *string, due *time.Time) (uint64, error) {
	tx, err := srv.DB.BeginTxx(ctxOrBackground(ctx), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	// lock the item row to prevent race
	var status string
	if err := tx.QueryRowContext(ctxOrBackground(ctx),
		`SELECT status FROM log_lab_equipment_master WHERE id=? FOR UPDATE`, itemID).Scan(&status); err != nil {
		return 0, err
	}
	if status != "available" {
		return 0, sql.ErrTxDone // replace with custom error if desired
	}

	res, err := tx.ExecContext(ctxOrBackground(ctx), `
		INSERT INTO log_lab_borrow_records
			(item_id, borrower_name, borrower_id, borrowed_at, due_at, status)
		VALUES (?,?,?,?,?, 'active')`,
		itemID, borrowerName, borrowerID, time.Now().UTC(), due,
	)
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctxOrBackground(ctx),
		`UPDATE log_lab_equipment_master SET status='borrowed' WHERE id=?`, itemID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return uint64(id), nil
}

func BorrowReturn(ctx context.Context, borrowID uint64) error {
	tx, err := srv.DB.BeginTxx(ctxOrBackground(ctx), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var itemID uint64
	if err := tx.QueryRowContext(ctxOrBackground(ctx),
		`SELECT item_id FROM log_lab_borrow_records WHERE id=? AND status='active' FOR UPDATE`, borrowID).Scan(&itemID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctxOrBackground(ctx),
		`UPDATE log_lab_borrow_records SET status='returned', returned_at=? WHERE id=?`,
		time.Now().UTC(), borrowID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctxOrBackground(ctx),
		`UPDATE log_lab_equipment_master SET status='available' WHERE id=?`, itemID); err != nil {
		return err
	}
	return tx.Commit()
}

// =========================
//Maintenance (log_lab_maintenance_records)
// =========================

func MaintCreate(ctx context.Context, m *models.MaintenanceRecord) (uint64, error) {
	if m.OpenedAt.IsZero() {
		m.OpenedAt = time.Now().UTC()
	}
	if m.Status == "" {
		m.Status = "open"
	}
	res, err := srv.DB.ExecContext(ctxOrBackground(ctx), `
		INSERT INTO log_lab_maintenance_records
			(item_id, title, description, priority, status, opened_at, closed_at, assigned_to)
		VALUES (?,?,?,?,?,?,?,?)`,
		m.ItemID, m.Title, m.Description, m.Priority, m.Status, m.OpenedAt, m.ClosedAt, m.AssignedTo,
	)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return uint64(id), nil
}

func MaintListOpen(ctx context.Context) ([]models.MaintenanceRecord, error) {
	var out []models.MaintenanceRecord
	err := srv.DB.SelectContext(ctxOrBackground(ctx), &out, `
		SELECT id, item_id, title, description, priority, status, opened_at, closed_at, assigned_to
		FROM log_lab_maintenance_records
		WHERE status <> 'closed'
		ORDER BY opened_at DESC`)
	return out, err
}

// =========================
// Calibration (log_lab_calibration_logs)
// =========================

func CalibrationLogCreate(ctx context.Context, c *models.CalibrationLog) (uint64, error) {
	if c.PerformedAt.IsZero() {
		c.PerformedAt = time.Now().UTC()
	}
	res, err := srv.DB.ExecContext(ctxOrBackground(ctx), `
		INSERT INTO log_lab_calibration_logs
			(item_id, performed_at, next_due, result, technician, notes)
		VALUES (?,?,?,?,?,?)`,
		c.ItemID, c.PerformedAt, c.NextDue, c.Result, c.Technician, c.Notes,
	)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return uint64(id), nil
}

func CalibrationLatestForItem(ctx context.Context, itemID uint64) (*models.CalibrationLog, error) {
	var c models.CalibrationLog
	err := srv.DB.GetContext(ctxOrBackground(ctx), &c, `
		SELECT id, item_id, performed_at, next_due, result, technician, notes
		FROM log_lab_calibration_logs
		WHERE item_id=?
		ORDER BY performed_at DESC
		LIMIT 1`, itemID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

// =========================
// Storage (log_lab_storage)
// =========================

func StorageUpsert(ctx context.Context, s *models.Storage) error {
	_, err := srv.DB.ExecContext(ctxOrBackground(ctx), `
		INSERT INTO log_lab_storage (item_id, location, quantity, updated_at)
		VALUES (?,?,?, NOW())
		ON DUPLICATE KEY UPDATE location=VALUES(location), quantity=VALUES(quantity), updated_at=NOW()`,
		s.ItemID, s.Location, s.Quantity,
	)
	return err
}

func StorageByItem(ctx context.Context, itemID uint64) ([]models.Storage, error) {
	var out []models.Storage
	err := srv.DB.SelectContext(ctxOrBackground(ctx), &out, `
		SELECT id, item_id, location, quantity, updated_at
		FROM log_lab_storage
		WHERE item_id=?
		ORDER BY updated_at DESC`, itemID)
	return out, err
}

// =========================
// Activity log (log_lab_activity_logs)
// =========================

func ActivityWrite(ctx context.Context, actor, action, target string, details interface{}) error {
	var d []byte
	if details != nil {
		if b, err := json.Marshal(details); err == nil {
			d = b
		}
	}
	_, err := srv.DB.ExecContext(ctxOrBackground(ctx), `
		INSERT INTO log_lab_activity_logs (actor, action, target, details, created_at)
		VALUES (?,?,?, ?, NOW())`,
		nullableStr(actor), action, nullableStr(target), d,
	)
	return err
}

// helpers

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
