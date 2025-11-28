package models

import "time"

// ----- users -----
type User struct {
	ID           uint64     `db:"id"            json:"id"`
	Username     string     `db:"username"      json:"username"`
	PasswordHash []byte     `db:"password_hash" json:"-"`
	FullName     string     `db:"full_name"     json:"full_name"`
	Email        string     `db:"email"         json:"email"`
	Role         string     `db:"role"          json:"role"`
	LastLogin    *time.Time `db:"last_login"    json:"last_login,omitempty"`
	CreatedAt    time.Time  `db:"create_at"     json:"create_at"`
}

// ----- log_lab_equipment_master -----
// Inventory master list
type Equipment struct {
	ID                int      `json:"id"                  db:"id"`
	SKU               string   `json:"sku"                 db:"sku"`
	Name              *string  `json:"name"                db:"name"`
	Description       *string  `json:"description"         db:"description"`
	Category          *string  `json:"category"            db:"category"`
	Location          *string  `json:"location"            db:"location"`
	Quantity          *int     `json:"quantity"            db:"quantity"`
	AvailableQuantity *int     `json:"available_quantity"  db:"available_quantity"`
	UnitCost          *float32 `json:"unit_cost"           db:"unit_cost"`
	Supplier          *string  `json:"supplier"            db:"supplier"`
	DatePurchased     *string  `json:"date_purchased"      db:"date_purchased"` // YYYY-MM-DD
	Status            *string  `json:"status"              db:"status"`
	CreatedAt         *string  `json:"create_at"           db:"create_at"` // YYYY-MM-DD HH:MM:SS
}

// ----- log_lab_borrow_records -----
type BorrowRecord struct {
	ID           uint64     `db:"id" json:"id"`
	ItemID       uint64     `db:"item_id" json:"item_id"`
	BorrowerName string     `db:"borrower_name" json:"borrower_name"`
	BorrowerID   *string    `db:"borrower_id" json:"borrower_id,omitempty"`
	BorrowedAt   time.Time  `db:"borrowed_at" json:"borrowed_at"`
	DueAt        *time.Time `db:"due_at" json:"due_at,omitempty"`
	ReturnedAt   *time.Time `db:"returned_at" json:"returned_at,omitempty"`
	Status       string     `db:"status" json:"status"` // active|returned|overdue
}

// ----- log_lab_maintenance_records -----
type MaintenanceRecord struct {
	ID          uint64     `db:"id" json:"id"`
	ItemID      uint64     `db:"item_id" json:"item_id"`
	Title       string     `db:"title" json:"title"`
	Description *string    `db:"description" json:"description,omitempty"`
	Priority    string     `db:"priority" json:"priority"` // low|medium|high|urgent
	Status      string     `db:"status" json:"status"`     // open|scheduled|in_progress|closed
	OpenedAt    time.Time  `db:"opened_at" json:"opened_at"`
	ClosedAt    *time.Time `db:"closed_at" json:"closed_at,omitempty"`
	AssignedTo  *string    `db:"assigned_to" json:"assigned_to,omitempty"`
}

// ----- log_lab_calibration_logs -----
type CalibrationLog struct {
	ID          uint64     `db:"id" json:"id"`
	ItemID      uint64     `db:"item_id" json:"item_id"`
	PerformedAt time.Time  `db:"performed_at" json:"performed_at"`
	NextDue     *time.Time `db:"next_due" json:"next_due,omitempty"`
	Result      *string    `db:"result" json:"result,omitempty"` // pass|fail|...
	Technician  *string    `db:"technician" json:"technician,omitempty"`
	Notes       *string    `db:"notes" json:"notes,omitempty"`
}

// ----- log_lab_storage -----
type Storage struct {
	ID        uint64    `db:"id" json:"id"`
	ItemID    uint64    `db:"item_id" json:"item_id"`
	Location  *string   `db:"location" json:"location,omitempty"`
	Quantity  int       `db:"quantity" json:"quantity"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ----- log_lab_activity_logs -----
type ActivityLog struct {
	ID        uint64    `db:"id" json:"id"`
	Actor     *string   `db:"actor" json:"actor,omitempty"`
	Action    string    `db:"action" json:"action"`
	Target    *string   `db:"target" json:"target,omitempty"`
	Details   []byte    `db:"details" json:"details,omitempty"` // JSON blob
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
