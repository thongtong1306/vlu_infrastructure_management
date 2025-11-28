package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web"
	beegoctx "github.com/beego/beego/v2/server/web/context"
	"github.com/go-sql-driver/mysql"
)

// ItemController serves equipment endpoints.
type ItemController struct{ web.Controller }

// request/response models
type addItemReq struct {
	SKU               string   `json:"sku"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	Category          string   `json:"category"`
	Location          string   `json:"location"`
	Quantity          *int     `json:"quantity"`
	AvailableQuantity *int     `json:"available_quantity"`
	UnitCost          *float64 `json:"unit_cost"`
	Supplier          string   `json:"supplier"`
	DatePurchased     string   `json:"date_purchased"` // YYYY-MM-DD
	Status            string   `json:"status"`
}

type itemRow struct {
	ID                int        `db:"id"                  json:"id"`
	SKU               string     `db:"sku"                 json:"sku"`
	Name              string     `db:"name"                json:"name"`
	Description       string     `db:"description"         json:"description"`
	ImageURL          *string    `db:"image_url"           json:"image_url,omitempty"`
	Category          string     `db:"category"            json:"category"`
	Location          string     `db:"location"            json:"location"`
	Quantity          int        `db:"quantity"            json:"quantity"`
	AvailableQuantity int        `db:"available_quantity"  json:"available_quantity"`
	UnitCost          float64    `db:"unit_cost"           json:"unit_cost"`
	Supplier          string     `db:"supplier"            json:"supplier"`
	DatePurchased     *time.Time `db:"date_purchased"      json:"date_purchased,omitempty"`
	Status            string     `db:"status"              json:"status"`
	CreateAt          time.Time  `db:"create_at"           json:"create_at"`
}

type borrowReq struct {
	SKU        string `json:"sku"`                   // identify by SKU ...
	ItemID     *int   `json:"item_id,omitempty"`     // ... or ItemID
	Quantity   int    `json:"quantity"`              // > 0
	ReturnDate string `json:"return_date,omitempty"` // optional YYYY-MM-DD
}

type borrowResp struct {
	ID              int        `json:"id"`
	UserID          int        `json:"user_id"`
	ItemID          int        `json:"item_id"`
	Quantity        int        `json:"quantity"`
	BorrowDate      time.Time  `json:"borrow_date"`
	ReturnDate      *time.Time `json:"return_date,omitempty"`
	ActualReturn    *time.Time `json:"actual_return_date,omitempty"`
	ConditionReturn *string    `json:"condition_on_return,omitempty"`
	Status          string     `json:"status"`
}

// ---------- POST /api/items ----------
func (c *ItemController) Add() {
	if !isAuthed(c.Ctx) {
		unauthorized(c)
		return
	}

	var in addItemReq
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&in); err != nil {
		badRequest(c, "invalid json")
		return
	}

	in.SKU = trim(in.SKU)
	in.Name = trim(in.Name)
	in.Description = trim(in.Description)
	in.Category = trim(in.Category)
	in.Location = trim(in.Location)
	in.Supplier = trim(in.Supplier)
	if in.Status == "" {
		in.Status = "active"
	}

	if in.SKU == "" || in.Name == "" {
		badRequest(c, "sku and name are required")
		return
	}
	qty := 0
	if in.Quantity != nil {
		qty = *in.Quantity
	}
	if qty < 0 {
		badRequest(c, "quantity must be >= 0")
		return
	}
	avail := qty
	if in.AvailableQuantity != nil {
		avail = *in.AvailableQuantity
	}
	if avail < 0 || avail > qty {
		badRequest(c, "available_quantity must be between 0 and quantity")
		return
	}
	unit := 0.0
	if in.UnitCost != nil {
		unit = *in.UnitCost
	}
	if unit < 0 {
		badRequest(c, "unit_cost must be >= 0")
		return
	}

	var datePtr *time.Time
	if in.DatePurchased != "" {
		dt, err := parseDateYMD(in.DatePurchased)
		if err != nil {
			badRequest(c, "date_purchased must be YYYY-MM-DD")
			return
		}
		datePtr = dt
	}

	const q = `
		INSERT INTO log_lab_equipment_master
		(name, description, category, image_url, location, quantity, available_quantity, unit_cost, supplier, date_purchased, sku, status, create_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?, NOW())
	`

	var dateVal interface{}
	if datePtr == nil {
		dateVal = sql.NullTime{Valid: false}
	} else {
		t := time.Date(datePtr.Year(), datePtr.Month(), datePtr.Day(), 0, 0, 0, 0, time.UTC)
		dateVal = t
	}

	res, err := srv.DB.Exec(q,
		in.Name, in.Description, in.Category, in.Location,
		qty, avail, unit, in.Supplier, dateVal, in.SKU, in.Status,
	)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			badRequest(c, "duplicate sku")
			return
		}
		serverError(c, "insert error")
		return
	}
	id64, _ := res.LastInsertId()
	id := int(id64)

	var out itemRow
	getQ := `
		SELECT id, sku, name, description, category, image_url, location, quantity, available_quantity, unit_cost, supplier,
		       date_purchased, status, create_at
		FROM log_lab_equipment_master WHERE id=? LIMIT 1
	`
	if err := srv.DB.Get(&out, getQ, id); err != nil {
		now := time.Now().UTC()
		out = itemRow{
			ID:                id,
			SKU:               in.SKU,
			Name:              in.Name,
			Description:       in.Description,
			Category:          in.Category,
			Location:          in.Location,
			Quantity:          qty,
			AvailableQuantity: avail,
			UnitCost:          unit,
			Supplier:          in.Supplier,
			Status:            in.Status,
			CreateAt:          now,
		}
		if datePtr != nil {
			out.DatePurchased = datePtr
		}
	}

	c.Ctx.Output.SetStatus(http.StatusCreated)
	_ = c.Ctx.Output.JSON(out, false, false)
}

// Optional stub to avoid missing-method panics if routed:
// GET /api/items?q=osc&limit=200&offset=0
// GET /api/items?q=&limit=200&offset=0
func (c *ItemController) GetAll() {
	if !isAuthed(c.Ctx) {
		unauthorized(c)
		return
	}

	q := strings.TrimSpace(c.GetString("q"))
	limit := 200
	if v := c.GetString("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}
	offset := 0
	if v := c.GetString("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	args := []interface{}{}
	sqlStr := `
		SELECT id, sku, name, description, category, image_url, location,
		       quantity, available_quantity, unit_cost, supplier,
		       date_purchased, status, create_at
		FROM log_lab_equipment_master
		WHERE 1=1`
	if q != "" {
		sqlStr += ` AND (name LIKE ? OR sku LIKE ? OR category LIKE ? OR location LIKE ?)`
		p := "%" + q + "%"
		args = append(args, p, p, p, p)
	}
	sqlStr += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	var rows []itemRow
	if err := srv.DB.Select(&rows, sqlStr, args...); err != nil {
		// helpful error for dev
		log.Printf("[items] query error: %v", err)
		serverError(c, "query error")
		return
	}
	_ = c.Ctx.Output.JSON(rows, false, false)
}

// ---------- BORROW ----------
func (c *ItemController) Borrow() {
	uid, ok := requireUserID(c.Ctx)
	if !ok {
		unauthorized(c)
		return
	}

	var in borrowReq
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&in); err != nil {
		badRequest(c, "invalid json")
		return
	}
	in.SKU = strings.TrimSpace(in.SKU)
	if in.Quantity <= 0 {
		badRequest(c, "quantity must be > 0")
		return
	}

	// resolve item
	var itemID int
	var name, sku string
	var avail int

	tx, err := srv.DB.Beginx()
	if err != nil {
		serverError(c, "begin tx error")
		return
	}
	defer func() { _ = tx.Rollback() }()

	if in.ItemID != nil {
		// lock row
		row := tx.QueryRowx(`SELECT id, sku, name, available_quantity FROM log_lab_equipment_master WHERE id=? FOR UPDATE`, *in.ItemID)
		if err := row.Scan(&itemID, &sku, &name, &avail); err != nil {
			badRequest(c, "item not found")
			return
		}
	} else if in.SKU != "" {
		row := tx.QueryRowx(`SELECT id, sku, name, available_quantity FROM log_lab_equipment_master WHERE sku=? FOR UPDATE`, in.SKU)
		if err := row.Scan(&itemID, &sku, &name, &avail); err != nil {
			badRequest(c, "item not found")
			return
		}
	} else {
		badRequest(c, "provide sku or item_id")
		return
	}

	if avail < in.Quantity {
		badRequest(c, fmt.Sprintf("not enough stock: available=%d", avail))
		return
	}

	// optional return date
	var returnDatePtr *time.Time
	if strings.TrimSpace(in.ReturnDate) != "" {
		dt, err := time.Parse("2006-01-02", in.ReturnDate)
		if err != nil {
			badRequest(c, "return_date must be YYYY-MM-DD")
			return
		}
		t := time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, time.UTC)
		returnDatePtr = &t
	}

	// update stock
	if _, err := tx.Exec(`UPDATE log_lab_equipment_master SET available_quantity = available_quantity - ? WHERE id=?`,
		in.Quantity, itemID); err != nil {
		serverError(c, "update stock error")
		return
	}

	// create borrow record
	res, err := tx.Exec(`
		INSERT INTO log_lab_borrow_records
		(user_id, item_id, quantity, borrow_date, return_date, actual_return_date, condition_on_return, status)
		VALUES (?,?,?, NOW(), ?, NULL, NULL, 'borrowed')
	`, uid, itemID, in.Quantity, returnDatePtr)
	if err != nil {
		// unique (user_id, item_id, borrow_date) very unlikely to collide, but handle anyway
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			badRequest(c, "duplicate borrow record")
			return
		}
		serverError(c, "insert borrow error")
		return
	}
	id64, _ := res.LastInsertId()

	// activity
	logActivityTX(tx.Tx, uid, fmt.Sprintf("Borrowed %s (%s) x%d", name, sku, in.Quantity))

	if err := tx.Commit(); err != nil {
		serverError(c, "commit error")
		return
	}

	now := time.Now().UTC()
	out := borrowResp{
		ID:         int(id64),
		UserID:     uid,
		ItemID:     itemID,
		Quantity:   in.Quantity,
		BorrowDate: now,
		Status:     "borrowed",
	}
	if returnDatePtr != nil {
		out.ReturnDate = returnDatePtr
	}
	c.Ctx.Output.SetStatus(http.StatusCreated)
	_ = c.Ctx.Output.JSON(out, false, false)
}

// ---------- RETURN ----------
type returnReq struct {
	BorrowID          *int   `json:"borrow_id,omitempty"` // preferred if multiple open
	SKU               string `json:"sku,omitempty"`       // or identify open record by sku + user
	ItemID            *int   `json:"item_id,omitempty"`
	Quantity          *int   `json:"quantity,omitempty"` // defaults to full
	ConditionOnReturn string `json:"condition_on_return,omitempty"`
	ReturnedAt        string `json:"returned_at,omitempty"` // YYYY-MM-DD (optional), default NOW()
}

func (c *ItemController) Return() {
	uid, ok := requireUserID(c.Ctx)
	if !ok {
		unauthorized(c)
		return
	}

	var in returnReq
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&in); err != nil {
		badRequest(c, "invalid json")
		return
	}
	in.SKU = strings.TrimSpace(in.SKU)

	tx, err := srv.DB.Beginx()
	if err != nil {
		serverError(c, "begin tx error")
		return
	}
	defer func() { _ = tx.Rollback() }()

	// locate open borrow record
	type rec struct {
		ID       int
		ItemID   int
		Qty      int
		ItemName string
		SKU      string
	}
	var r rec

	if in.BorrowID != nil {
		row := tx.QueryRowx(`
			SELECT br.id, br.item_id, br.quantity, em.name, em.sku
			FROM log_lab_borrow_records br
			JOIN log_lab_equipment_master em ON em.id = br.item_id
			WHERE br.id=? AND br.user_id=? AND br.actual_return_date IS NULL AND br.status <> 'returned'
			LIMIT 1`, *in.BorrowID, uid)
		if err := row.Scan(&r.ID, &r.ItemID, &r.Qty, &r.ItemName, &r.SKU); err != nil {
			badRequest(c, "open borrow not found for this id")
			return
		}
	} else {
		// find by user + (sku OR item_id)
		cond := `br.user_id=? AND br.actual_return_date IS NULL AND br.status <> 'returned'`
		var args []interface{}
		args = append(args, uid)
		if in.ItemID != nil {
			cond += ` AND br.item_id=?`
			args = append(args, *in.ItemID)
		} else if in.SKU != "" {
			cond += ` AND em.sku=?`
			args = append(args, in.SKU)
		} else {
			badRequest(c, "provide borrow_id or sku/item_id")
			return
		}

		// ensure uniqueness (if multiple, ask for borrow_id)
		rows, err := tx.Queryx(`
			SELECT br.id, br.item_id, br.quantity, em.name, em.sku
			FROM log_lab_borrow_records br
			JOIN log_lab_equipment_master em ON em.id = br.item_id
			WHERE `+cond, args...)
		if err != nil {
			serverError(c, "query error")
			return
		}
		defer rows.Close()
		found := 0
		for rows.Next() {
			if err := rows.Scan(&r.ID, &r.ItemID, &r.Qty, &r.ItemName, &r.SKU); err != nil {
				serverError(c, "scan error")
				return
			}
			found++
			if found > 1 {
				badRequest(c, "multiple open borrows; specify borrow_id")
				return
			}
		}
		if found == 0 {
			badRequest(c, "open borrow not found")
			return
		}
	}

	// quantity: only support full return for now
	qty := r.Qty
	if in.Quantity != nil && *in.Quantity != r.Qty {
		badRequest(c, "partial returns not supported yet (send full quantity)")
		return
	}

	// returned at
	var retAt *time.Time
	if strings.TrimSpace(in.ReturnedAt) != "" {
		dt, err := time.Parse("2006-01-02", in.ReturnedAt)
		if err != nil {
			badRequest(c, "returned_at must be YYYY-MM-DD")
			return
		}
		t := time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, time.UTC)
		retAt = &t
	}

	// update record
	if _, err := tx.Exec(`
		UPDATE log_lab_borrow_records
		SET actual_return_date = IFNULL(?, NOW()),
		    condition_on_return = NULLIF(?, ''),
		    status = 'returned'
		WHERE id=? AND actual_return_date IS NULL
	`, retAt, in.ConditionOnReturn, r.ID); err != nil {
		serverError(c, "update borrow error")
		return
	}

	// restore stock
	if _, err := tx.Exec(`UPDATE log_lab_equipment_master SET available_quantity = available_quantity + ? WHERE id=?`,
		qty, r.ItemID); err != nil {
		serverError(c, "restore stock error")
		return
	}

	// activity
	logActivityTX(tx.Tx, uid, fmt.Sprintf("Returned %s (%s) x%d", r.ItemName, r.SKU, qty))

	if err := tx.Commit(); err != nil {
		serverError(c, "commit error")
		return
	}

	_ = c.Ctx.Output.JSON(map[string]string{"status": "ok"}, false, false)
}

func (c *ItemController) UpdateImageURL() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		serverError(c, "invalid id")
		return
	}

	var in struct {
		ImageURL string `json:"image_url"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil || strings.TrimSpace(in.ImageURL) == "" {
		serverError(c, "invalid body")
		return
	}
	if _, err := srv.DB.Exec(`UPDATE log_lab_equipment_master SET image_url=? WHERE id=?`, strings.TrimSpace(in.ImageURL), id); err != nil {
		serverError(c, "update failed")
		return
	}
	_ = c.Ctx.Output.JSON(map[string]interface{}{"ok": true}, false, false)
}

// ---------- helpers ----------
func trim(s string) string { return strings.TrimSpace(s) }

func parseDateYMD(s string) (*time.Time, error) {
	s = trim(s)
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func badRequest(c *ItemController, msg string) {
	c.Ctx.Output.SetStatus(http.StatusBadRequest)
	_ = c.Ctx.Output.JSON(map[string]string{"error": msg}, false, false)
}

func serverError(c *ItemController, msg string) {
	c.Ctx.Output.SetStatus(http.StatusInternalServerError)
	_ = c.Ctx.Output.JSON(map[string]string{"error": msg}, false, false)
}

func unauthorized(c *ItemController) {
	c.Ctx.Output.SetStatus(http.StatusUnauthorized)
	_ = c.Ctx.Output.JSON(map[string]string{"error": "missing or invalid auth"}, false, false)
}

// Use the correct Beego v2 context type here:
func isAuthed(ctx *beegoctx.Context) bool {
	// If your auth middleware sets user_id:
	if _, ok := ctx.Input.GetData("user_id").(int); ok {
		return true
	}
	// Fallback: accept presence of a Bearer header (change to strict as needed)
	if auth := ctx.Input.Header("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return true
	}
	return false
}

func (c *ItemController) GetOne() {
	if !isAuthed(c.Ctx) {
		unauthorized(c)
		return
	}
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		serverError(c, "invalid id")
		return
	}

	var row itemRow
	err = srv.DB.Get(&row, `
		SELECT id, sku, name, description, image_url, category, location,
		       quantity, available_quantity, unit_cost, supplier,
		       date_purchased, status, create_at
		FROM log_lab_equipment_master
		WHERE id=? LIMIT 1`, id)
	if err != nil {
		if err == sql.ErrNoRows {
			jsonErr(c.Ctx, http.StatusNotFound, "not found")
			return
		}
		serverError(c, "query error")
		return
	}
	_ = c.Ctx.Output.JSON(row, false, false)
}

// requireUserID returns a user id by checking context data first,
// then falling back to a lookup by userEmail (set by SessionAuthFilter).
func requireUserID(ctx *beegoctx.Context) (int, bool) {
	if v, ok := ctx.Input.GetData("user_id").(int); ok && v > 0 {
		return v, true
	}
	if email, ok := ctx.Input.GetData("userEmail").(string); ok && email != "" {
		var id int
		if err := srv.DB.Get(&id, "SELECT id FROM "+usersTable+" WHERE email=? LIMIT 1", email); err == nil && id > 0 {
			ctx.Input.SetData("user_id", id) // cache for downstream handlers
			return id, true
		}
	}
	return 0, false
}

func logActivityTX(tx *sql.Tx, userID int, action string) {
	_, _ = tx.Exec(`INSERT INTO log_lab_activity_logs (user_id, action, timestamp) VALUES (?,?, NOW())`,
		userID, action)
}
