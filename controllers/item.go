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

// Add creates a new equipment row.
func (c *ItemController) Add() {
	// tiny local helpers that write straight to the Beego context
	ok := func(status int, data interface{}) {
		c.Ctx.Output.SetStatus(status)
		_ = c.Ctx.Output.JSON(data, false, false)
	}
	errf := func(status int, msg string) {
		c.Ctx.Output.SetStatus(status)
		_ = c.Ctx.Output.JSON(map[string]interface{}{"ok": false, "error": msg}, false, false)
	}

	// ---- auth ----
	if !isAuthed(c.Ctx) {
		errf(401, "unauthorized")
		return
	}

	// ---- payload ----
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
		DatePurchased     string   `json:"date_purchased"` // "YYYY-MM-DD" or ""
		Status            string   `json:"status"`
		ImageURL          *string  `json:"image_url"` // optional
	}

	var in addItemReq
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&in); err != nil {
		errf(400, "invalid json")
		return
	}

	// ---- sanitize / defaults ----
	in.SKU = strings.TrimSpace(in.SKU)
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.Category = strings.TrimSpace(in.Category)
	in.Location = strings.TrimSpace(in.Location)
	in.Supplier = strings.TrimSpace(in.Supplier)
	in.Status = strings.TrimSpace(in.Status)
	if in.Status == "" {
		in.Status = "active"
	}
	if in.Name == "" {
		errf(400, "name is required")
		return
	}

	qty := 0
	if in.Quantity != nil && *in.Quantity > 0 {
		qty = *in.Quantity
	}
	avail := qty
	if in.AvailableQuantity != nil {
		avail = *in.AvailableQuantity
		if avail < 0 {
			avail = 0
		}
	}

	var unit interface{} = nil
	if in.UnitCost != nil {
		unit = *in.UnitCost
	}

	var image interface{} = nil
	if in.ImageURL != nil {
		if trimmed := strings.TrimSpace(*in.ImageURL); trimmed != "" {
			image = trimmed
		}
	}

	var dateVal interface{} = nil // NULL if empty
	if d := strings.TrimSpace(in.DatePurchased); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			errf(400, "date_purchased must be YYYY-MM-DD")
			return
		}
		dateVal = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}

	// ---- INSERT (column order must match placeholders) ----
	const insertSQL = `
		INSERT INTO log_lab_equipment_master
		  (name, description, category, image_url, location,
		   quantity, available_quantity, unit_cost, supplier,
		   date_purchased, sku, status, create_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?, ?, NOW())
	`

	res, err := srv.DB.Exec(insertSQL,
		in.Name,        // name
		in.Description, // description
		in.Category,    // category
		image,          // image_url
		in.Location,    // location
		qty,            // quantity
		avail,          // available_quantity
		unit,           // unit_cost
		in.Supplier,    // supplier
		dateVal,        // date_purchased
		in.SKU,         // sku
		in.Status,      // status
	)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			errf(400, "duplicate key (likely sku)")
			return
		}
		errf(500, "insert error: "+err.Error())
		return
	}

	newID, _ := res.LastInsertId()

	// ---- return created row ----
	row := map[string]interface{}{}
	const getSQL = `
		SELECT id, sku, name, description, image_url, category, location,
		       quantity, available_quantity, unit_cost, supplier,
		       date_purchased, status, create_at
		FROM log_lab_equipment_master
		WHERE id = ? LIMIT 1
	`
	r := srv.DB.QueryRowx(getSQL, newID)
	if r != nil {
		if err := r.MapScan(row); err != nil && !errors.Is(err, sql.ErrNoRows) {
			ok(200, map[string]interface{}{"ok": true, "id": newID})
			return
		}
	}
	if len(row) == 0 {
		ok(200, map[string]interface{}{"ok": true, "id": newID})
		return
	}
	ok(200, map[string]interface{}{"ok": true, "data": row})
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

// uses return_date (or due_date) and updates stock.
func (c *ItemController) Borrow() {
	// local JSON helpers
	ok := func(status int, data interface{}) {
		c.Ctx.Output.SetStatus(status)
		_ = c.Ctx.Output.JSON(data, false, false)
	}
	errf := func(status int, msg string) {
		c.Ctx.Output.SetStatus(status)
		_ = c.Ctx.Output.JSON(map[string]any{"ok": false, "error": msg}, false, false)
	}

	// payload
	type borrowReq struct {
		UserID   *int64 `json:"user_id"`     // optional; fallback to context
		ItemID   *int64 `json:"item_id"`     // optional if sku provided
		SKU      string `json:"sku"`         // optional if item_id provided
		Quantity int    `json:"quantity"`    // required (>0)
		Return   string `json:"return_date"` // "YYYY-MM-DD" (optional)
		Due      string `json:"due_date"`    // alias to return_date
	}

	var in borrowReq
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&in); err != nil {
		errf(400, "invalid json")
		return
	}
	if in.Quantity <= 0 {
		errf(400, "quantity must be > 0")
		return
	}

	// Resolve user_id: prefer body, else context
	resolveUserID := func() (int64, bool) {
		if in.UserID != nil && *in.UserID > 0 {
			return *in.UserID, true
		}
		if v := c.Ctx.Input.GetData("user_id"); v != nil {
			switch t := v.(type) {
			case int64:
				return t, true
			case int:
				return int64(t), true
			case float64:
				return int64(t), true
			case string:
				if id, e := strconv.ParseInt(t, 10, 64); e == nil {
					return id, true
				}
			}
		}
		return 0, false
	}
	userID, okUID := resolveUserID()
	if !okUID || userID <= 0 {
		errf(401, "user id missing")
		return
	}

	// Resolve item_id: accept either item_id or sku
	var itemID int64
	if in.ItemID != nil && *in.ItemID > 0 {
		itemID = *in.ItemID
	} else if s := in.SKU; s != "" {
		if err := srv.DB.Get(&itemID, `SELECT id FROM log_lab_equipment_master WHERE sku = ? LIMIT 1`, s); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				errf(404, "item not found by sku")
				return
			}
			errf(500, "lookup sku: "+err.Error())
			return
		}
	} else {
		errf(400, "provide item_id or sku")
		return
	}

	// Parse return date (either return_date or due_date)
	var returnDate interface{} = nil
	if d := strings.TrimSpace(in.Return); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			errf(400, "return_date must be YYYY-MM-DD")
			return
		}
		returnDate = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	} else if d := strings.TrimSpace(in.Due); d != "" {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			errf(400, "due_date must be YYYY-MM-DD")
			return
		}
		returnDate = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	}

	tx, err := srv.DB.Beginx()
	if err != nil {
		errf(500, "tx begin: "+err.Error())
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Lock & check stock
	var avail int
	if err = tx.Get(&avail, `SELECT available_quantity FROM log_lab_equipment_master WHERE id=? FOR UPDATE`, itemID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errf(404, "item not found")
			return
		}
		errf(500, "query item: "+err.Error())
		return
	}
	if avail < in.Quantity {
		errf(400, "not enough stock")
		return
	}

	// Insert borrow row (matches your columns)
	res, err := tx.Exec(`
		INSERT INTO log_lab_borrow_records
		  (user_id, item_id, quantity, borrow_date, return_date, actual_return_date, condition_on_return, status)
		VALUES
		  (      ?,       ?,        ?,        NOW(),         ?,               NULL,                NULL, 'borrowed')
	`, userID, itemID, in.Quantity, returnDate)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1452 {
			errf(400, "invalid user_id or item_id")
			return
		}
		errf(500, "insert borrow: "+err.Error())
		return
	}
	borrowID, _ := res.LastInsertId()

	// Update stock
	if _, err = tx.Exec(`UPDATE log_lab_equipment_master SET available_quantity = available_quantity - ? WHERE id=?`,
		in.Quantity, itemID); err != nil {
		errf(500, "update stock: "+err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		errf(500, "commit: "+err.Error())
		return
	}

	ok(200, map[string]any{
		"ok":       true,
		"id":       borrowID, // matches your UI message
		"item_id":  itemID,
		"quantity": in.Quantity,
	})
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
	// small helpers
	bad := func(msg string) { badRequest(c, msg) }
	serr := func(msg string) { serverError(c, msg) }

	// ---- payload ----
	type returnReq struct {
		UserID            *int64 `json:"user_id,omitempty"`   // optional; fallback to context
		BorrowID          *int   `json:"borrow_id,omitempty"` // preferred
		SKU               string `json:"sku,omitempty"`       // or identify by sku...
		ItemID            *int   `json:"item_id,omitempty"`   // ...or item id
		Quantity          *int   `json:"quantity,omitempty"`  // full only (for now)
		ConditionOnReturn string `json:"condition_on_return,omitempty"`
		ReturnedAt        string `json:"returned_at,omitempty"` // YYYY-MM-DD (optional)
	}
	var in returnReq
	if err := json.NewDecoder(c.Ctx.Request.Body).Decode(&in); err != nil {
		bad("invalid json")
		return
	}
	in.SKU = strings.TrimSpace(in.SKU)

	// ---- resolve user_id (body > context) ----
	resolveUID := func() (int64, bool) {
		if in.UserID != nil && *in.UserID > 0 {
			return *in.UserID, true
		}
		if v := c.Ctx.Input.GetData("user_id"); v != nil {
			switch t := v.(type) {
			case int64:
				return t, true
			case int:
				return int64(t), true
			case float64:
				return int64(t), true
			case string:
				if id, e := strconv.ParseInt(t, 10, 64); e == nil {
					return id, true
				}
			}
		}
		return 0, false
	}
	uid64, ok := resolveUID()
	if !ok || uid64 <= 0 {
		unauthorized(c)
		return
	}
	uid := int(uid64)

	tx, err := srv.DB.Beginx()
	if err != nil {
		serr("begin tx error")
		return
	}
	defer func() { _ = tx.Rollback() }()

	// ---- locate open borrow record ----
	type rec struct {
		ID     int
		ItemID int
		Qty    int
		Name   string
		SKU    string
	}
	var r rec

	if in.BorrowID != nil {
		row := tx.QueryRowx(`
			SELECT br.id, br.item_id, br.quantity, em.name, em.sku
			FROM log_lab_borrow_recordsw br
			JOIN log_lab_equipment_master em ON em.id = br.item_id
			WHERE br.id=? AND br.user_id=? AND br.actual_return_date IS NULL AND br.status <> 'returned'
			LIMIT 1`, *in.BorrowID, uid)
		if err := row.Scan(&r.ID, &r.ItemID, &r.Qty, &r.Name, &r.SKU); err != nil {
			bad("open borrow not found for this id")
			return
		}
	} else {
		cond := `br.user_id=? AND br.actual_return_date IS NULL AND br.status <> 'returned'`
		args := []interface{}{uid}
		if in.ItemID != nil {
			cond += ` AND br.item_id=?`
			args = append(args, *in.ItemID)
		} else if in.SKU != "" {
			cond += ` AND em.sku=?`
			args = append(args, in.SKU)
		} else {
			bad("provide borrow_id or sku/item_id")
			return
		}

		rows, err := tx.Queryx(`
			SELECT br.id, br.item_id, br.quantity, em.name, em.sku
			FROM log_lab_borrow_records br
			JOIN log_lab_equipment_master em ON em.id = br.item_id
			WHERE `+cond, args...)
		if err != nil {
			serr("query error")
			return
		}
		defer rows.Close()
		found := 0
		for rows.Next() {
			if err := rows.Scan(&r.ID, &r.ItemID, &r.Qty, &r.Name, &r.SKU); err != nil {
				serr("scan error")
				return
			}
			found++
			if found > 1 {
				bad("multiple open borrows; specify borrow_id")
				return
			}
		}
		if found == 0 {
			bad("open borrow not found")
			return
		}
	}

	// ---- quantity: full only (for now) ----
	qty := r.Qty
	if in.Quantity != nil && *in.Quantity != r.Qty {
		bad("partial returns not supported yet (send full quantity)")
		return
	}

	// ---- returned_at ----
	var retAt *time.Time
	if s := strings.TrimSpace(in.ReturnedAt); s != "" {
		dt, err := time.Parse("2006-01-02", s)
		if err != nil {
			bad("returned_at must be YYYY-MM-DD")
			return
		}
		t := time.Date(dt.Year(), dt.Month(), dt.Day(), 0, 0, 0, 0, time.UTC)
		retAt = &t
	}

	// ---- update borrow (use your table: log_lab_borrow) ----
	if _, err := tx.Exec(`
		UPDATE log_lab_borrow_records
		SET actual_return_date = IFNULL(?, NOW()),
		    condition_on_return = NULLIF(?, ''),
		    status = 'returned'
		WHERE id=? AND actual_return_date IS NULL
	`, retAt, in.ConditionOnReturn, r.ID); err != nil {
		serr("update borrow error")
		return
	}

	// ---- restore stock ----
	if _, err := tx.Exec(`UPDATE log_lab_equipment_master SET available_quantity = available_quantity + ? WHERE id=?`,
		qty, r.ItemID); err != nil {
		serr("restore stock error")
		return
	}

	// activity log
	logActivityTX(tx.Tx, uid, fmt.Sprintf("Returned %s (%s) x%d", r.Name, r.SKU, qty))

	if err := tx.Commit(); err != nil {
		serr("commit error")
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
