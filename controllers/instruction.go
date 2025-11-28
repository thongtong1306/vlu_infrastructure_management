package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web"
)

const (
	equipMasterTable       = "log_lab_equipment_master"
	equipInstructionsTable = "log_lab_equipment_instructions"
)

type Instruction struct {
	ID        int64      `db:"id"         json:"id"`
	ItemID    int64      `db:"item_id"    json:"item_id"`
	Title     string     `db:"title"      json:"title"`
	Body      string     `db:"body"       json:"body"`
	ImageURL  *string    `db:"image_url"  json:"image_url,omitempty"`
	CreatedAt *time.Time `db:"created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
}

type InstructionController struct{ web.Controller }

// GET /api/instructions?item_id=123
func (c *InstructionController) GetByItem() {
	itemIDStr := strings.TrimSpace(c.GetString("item_id"))
	if itemIDStr == "" {
		jsonErr(c.Ctx, http.StatusBadRequest, "item_id is required")
		return
	}
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil || itemID <= 0 {
		jsonErr(c.Ctx, http.StatusBadRequest, "invalid item_id")
		return
	}

	s := GetServer()
	var out []struct {
		ID        int64      `db:"id"         json:"id"`
		ItemID    int64      `db:"item_id"    json:"item_id"`
		Title     string     `db:"title"      json:"title"`
		CreatedAt *time.Time `db:"created_at" json:"created_at,omitempty"`
	}
	q := "SELECT id, item_id, title, created_at FROM " + equipInstructionsTable + " WHERE item_id=? ORDER BY id DESC"
	if err := s.DB.Select(&out, q, itemID); err != nil {
		jsonErr(c.Ctx, http.StatusInternalServerError, "query error")
		return
	}
	jsonOK(c.Ctx, out)
}

// GET /api/instructions/:id
func (c *InstructionController) GetOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		jsonErr(c.Ctx, http.StatusBadRequest, "invalid id")
		return
	}
	s := GetServer()
	var row Instruction
	q := "SELECT id, item_id, title, body, image_url, created_at, updated_at FROM " + equipInstructionsTable + " WHERE id=? LIMIT 1"
	if err := s.DB.Get(&row, q, id); err != nil {
		if err == sql.ErrNoRows {
			jsonErr(c.Ctx, http.StatusNotFound, "not found")
		} else {
			jsonErr(c.Ctx, http.StatusInternalServerError, "query error")
		}
		return
	}
	jsonOK(c.Ctx, row)
}

// POST /api/instructions
// { "item_id": 123, "title": "...", "body": "...", "image_url": "https://..." }
func (c *InstructionController) Add() {
	var in struct {
		ItemID   int64   `json:"item_id"`
		Title    string  `json:"title"`
		Body     string  `json:"body"`
		ImageURL *string `json:"image_url"`
	}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &in); err != nil {
		jsonErr(c.Ctx, http.StatusBadRequest, "invalid json")
		return
	}
	if in.ItemID <= 0 || strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Body) == "" {
		jsonErr(c.Ctx, http.StatusBadRequest, "item_id, title, body are required")
		return
	}

	s := GetServer()
	// optional existence check
	var exists int
	_ = s.DB.Get(&exists, "SELECT 1 FROM "+equipMasterTable+" WHERE id=? LIMIT 1", in.ItemID)
	if exists != 1 {
		jsonErr(c.Ctx, http.StatusBadRequest, "equipment not found")
		return
	}

	res, err := s.DB.Exec(
		"INSERT INTO "+equipInstructionsTable+" (item_id, title, body, image_url) VALUES (?,?,?,?)",
		in.ItemID, strings.TrimSpace(in.Title), strings.TrimSpace(in.Body), in.ImageURL,
	)
	if err != nil {
		jsonErr(c.Ctx, http.StatusInternalServerError, "insert error")
		return
	}
	id, _ := res.LastInsertId()
	jsonOK(c.Ctx, map[string]interface{}{"ok": true, "id": id})
}
